import http from "k6/http";
import { check, sleep } from "k6";
import { Counter, Rate, Trend } from "k6/metrics";

const postgresResponses = new Counter("postgres_responses");
const redisResponses = new Counter("redis_responses");
const unexpectedSources = new Counter("unexpected_sources");

const postgresSuccess = new Rate("postgres_search_success");
const redisSuccess = new Rate("redis_search_success");

const postgresLatency = new Trend("postgres_search_latency", true);
const redisLatency = new Trend("redis_search_latency", true);

const baseUrl =
    __ENV.BASE_URL || "http://host.docker.internal:8080";

const token = __ENV.TOKEN;

const cachedSearchURL =
    `${baseUrl}/api/search` +
    "?category_id=7" +
    "&min_price=100000" +
    "&max_price=1000000" +
    "&limit=5";

export const options = {
    scenarios: {
        postgres_path: {
            executor: "constant-vus",
            exec: "postgresPath",
            vus: 10,
            duration: "30s",
            startTime: "0s",
        },

        redis_path: {
            executor: "constant-vus",
            exec: "redisPath",
            vus: 10,
            duration: "30s",
            startTime: "35s",
        },
    },

    thresholds: {
        checks: ["rate>0.99"],
        http_req_failed: ["rate<0.01"],

        postgres_search_success: ["rate>0.99"],
        redis_search_success: ["rate>0.99"],

        postgres_search_latency: [
            "p(95)<500",
            "p(99)<1000",
        ],

        redis_search_latency: [
            "p(95)<200",
            "p(99)<500",
        ],

        "http_req_duration{path:postgres}": [
            "p(95)<500",
            "p(99)<1000",
        ],

        "http_req_duration{path:redis}": [
            "p(95)<200",
            "p(99)<500",
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

    const warmupResponse = http.get(cachedSearchURL, {
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
            `Cache warmup failed with status ${warmupResponse.status}: ` +
            warmupResponse.body
        );
    }

    console.log("Gateway healthy and Redis search key warmed");

    return {
        runSeed: Date.now(),
    };
}

export function postgresPath(data) {
    /*
     * Each request gets a different max_price.
     *
     * Because the cache key contains the search filters,
     * every request uses a new Redis key and therefore
     * reaches PostgreSQL.
     */
    const uniqueValue =
        data.runSeed +
        (__VU * 1000000) +
        __ITER;

    const uniqueSearchURL =
        `${baseUrl}/api/search` +
        "?category_id=7" +
        "&min_price=100000" +
        `&max_price=${1000000 + uniqueValue}` +
        "&limit=5";

    const response = http.get(uniqueSearchURL, {
        headers: {
            Authorization: `Bearer ${token}`,
        },

        tags: {
            endpoint: "search",
            path: "postgres",
            name: "GET /api/search PostgreSQL",
        },
    });

    postgresLatency.add(response.timings.duration);

    let body = {};

    try {
        body = response.json();
    } catch (error) {
        body = {};
    }

    const succeeded =
        response.status === 200 &&
        body.source === "postgres";

    postgresSuccess.add(succeeded);

    check(response, {
        "PostgreSQL search status is 200": (res) =>
            res.status === 200,

        "PostgreSQL response contains results": () =>
            Array.isArray(body.results),

        "PostgreSQL response contains count": () =>
            typeof body.count === "number",

        "PostgreSQL source is correct": () =>
            body.source === "postgres",
    });

    if (body.source === "postgres") {
        postgresResponses.add(1);
    } else {
        unexpectedSources.add(1);
    }

    sleep(0.2);
}

export function redisPath() {
    /*
     * Every request uses exactly the same filters.
     *
     * The setup function warmed this cache key, so these
     * requests should be served by Redis.
     */
    const response = http.get(cachedSearchURL, {
        headers: {
            Authorization: `Bearer ${token}`,
        },

        tags: {
            endpoint: "search",
            path: "redis",
            name: "GET /api/search Redis",
        },
    });

    redisLatency.add(response.timings.duration);

    let body = {};

    try {
        body = response.json();
    } catch (error) {
        body = {};
    }

    const succeeded =
        response.status === 200 &&
        body.source === "redis_cache";

    redisSuccess.add(succeeded);

    check(response, {
        "Redis search status is 200": (res) =>
            res.status === 200,

        "Redis response contains results": () =>
            Array.isArray(body.results),

        "Redis response contains count": () =>
            typeof body.count === "number",

        "Redis source is correct": () =>
            body.source === "redis_cache",
    });

    if (body.source === "redis_cache") {
        redisResponses.add(1);
    } else {
        unexpectedSources.add(1);
    }

    sleep(0.2);
}