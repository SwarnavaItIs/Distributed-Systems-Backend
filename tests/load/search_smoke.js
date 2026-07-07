import http from "k6/http";
import { check, sleep } from "k6";
import { Counter } from "k6/metrics";

const cacheHits = new Counter("search_cache_hits");
const postgresHits = new Counter("search_postgres_hits");

export const options = {
    vus: 1,
    iterations: 5,

    thresholds: {
        checks: ["rate==1"],
        http_req_failed: ["rate==0"],
        "http_req_duration{endpoint:search}": ["p(95)<500"],
    },
};

const baseUrl =
    __ENV.BASE_URL || "http://host.docker.internal:8080";

const token = __ENV.TOKEN;

export function setup() {
    if (!token) {
        throw new Error("TOKEN environment variable is required");
    }

    const response = http.get(`${baseUrl}/health`, {
        tags: {
            endpoint: "gateway_health",
            name: "GET /health",
        },
    });

    if (response.status !== 200) {
        throw new Error(
            `Gateway health check failed with status ${response.status}`
        );
    }

    return {};
}

export default function () {
    const url =
        `${baseUrl}/api/search` +
        "?category_id=7" +
        "&min_price=100000" +
        "&max_price=1000000" +
        "&limit=5";

    const response = http.get(url, {
        headers: {
            Authorization: `Bearer ${token}`,
        },

        tags: {
            endpoint: "search",
            name: "GET /api/search",
        },
    });

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

        "response contains count": () =>
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

    sleep(1);
}