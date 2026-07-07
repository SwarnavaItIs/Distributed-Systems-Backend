import http from "k6/http";
import { check, sleep } from "k6";
import { Counter, Rate, Trend } from "k6/metrics";

const cacheHits = new Counter("search_cache_hits");
const postgresHits = new Counter("search_postgres_hits");
const rateLimitedResponses = new Counter("rate_limited_responses");
const successfulSearches = new Rate("successful_searches");
const searchLatency = new Trend("search_latency", true);

const baseUrl =
    __ENV.BASE_URL || "http://host.docker.internal:8080";

const token = __ENV.TOKEN;

const searchURL =
    `${baseUrl}/api/search` +
    "?category_id=7" +
    "&min_price=100000" +
    "&max_price=1000000" +
    "&limit=5";

export const options = {
    scenarios: {
        search_load: {
            executor: "ramping-vus",
            startVUs: 0,

            stages: [
                { duration: "10s", target: 5 },
                { duration: "20s", target: 10 },
                { duration: "30s", target: 20 },
                { duration: "20s", target: 20 },
                { duration: "10s", target: 0 },
            ],

            gracefulRampDown: "5s",
        },
    },

    thresholds: {
        checks: ["rate>0.99"],
        successful_searches: ["rate>0.99"],
        http_req_failed: ["rate<0.01"],

        "http_req_duration{endpoint:search}": [
            "p(95)<500",
            "p(99)<1000",
        ],
    },
};

export function setup() {
    if (!token) {
        throw new Error("TOKEN environment variable is required");
    }

    const healthResponse = http.get(`${baseUrl}/health`, {
        tags: {
            endpoint: "gateway_health",
            name: "GET /health",
        },
    });

    if (healthResponse.status !== 200) {
        throw new Error(
            `Gateway health check failed with status ${healthResponse.status}`
        );
    }

    // Warm the Redis search cache before load begins.
    const warmupResponse = http.get(searchURL, {
        headers: {
            Authorization: `Bearer ${token}`,
        },

        tags: {
            endpoint: "search_warmup",
            name: "GET /api/search warmup",
        },
    });

    if (warmupResponse.status !== 200) {
        throw new Error(
            `Search warmup failed with status ${warmupResponse.status}: ` +
            warmupResponse.body
        );
    }

    console.log("Gateway healthy and search cache warmed");

    return {};
}

export default function () {
    const response = http.get(searchURL, {
        headers: {
            Authorization: `Bearer ${token}`,
        },

        tags: {
            endpoint: "search",
            name: "GET /api/search",
        },
    });

    searchLatency.add(response.timings.duration);

    if (response.status === 429) {
        rateLimitedResponses.add(1);
    }

    successfulSearches.add(response.status === 200);

    let responseBody = {};

    try {
        responseBody = response.json();
    } catch (error) {
        responseBody = {};
    }

    check(response, {
        "search status is 200": (res) => res.status === 200,

        "response contains results": () =>
            Array.isArray(responseBody.results),

        "response contains numeric count": () =>
            typeof responseBody.count === "number",

        "source is postgres or redis": () =>
            responseBody.source === "postgres" ||
            responseBody.source === "redis_cache",
    });

    if (responseBody.source === "postgres") {
        postgresHits.add(1);
    }

    if (responseBody.source === "redis_cache") {
        cacheHits.add(1);
    }

    sleep(0.2);
}