import http from "k6/http";
import { check } from "k6";
import { Counter } from "k6/metrics";

const allowedResponses = new Counter("allowed_responses");
const blockedResponses = new Counter("blocked_responses");
const unexpectedResponses = new Counter("unexpected_responses");

const baseUrl =
    __ENV.BASE_URL || "http://host.docker.internal:8080";

const token = __ENV.TOKEN;

/*
 * Both 200 and 429 are expected outcomes in this test.
 *
 * Without this callback, k6 would count the intentional
 * 429 responses as failed HTTP requests.
 */
http.setResponseCallback(
    http.expectedStatuses(200, 429)
);

export const options = {
    scenarios: {
        rate_limit_burst: {
            executor: "shared-iterations",
            vus: 50,
            iterations: 50,
            maxDuration: "20s",
        },
    },

    thresholds: {
        checks: ["rate==1"],
        http_req_failed: ["rate==0"],

        allowed_responses: ["count==20"],
        blocked_responses: ["count==30"],
        unexpected_responses: ["count==0"],
    },
};

export function setup() {
    if (!token) {
        throw new Error("TOKEN environment variable is required");
    }

    const healthResponse = http.get(`${baseUrl}/health`);

    if (healthResponse.status !== 200) {
        throw new Error(
            `Gateway health check failed with status ` +
            `${healthResponse.status}`
        );
    }

    console.log("Gateway healthy; beginning rate-limit burst");

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
            endpoint: "rate_limited_search",
            name: "GET /api/search rate-limit burst",
        },
    });

    let body = {};

    try {
        body = response.json();
    } catch (error) {
        body = {};
    }

    const allowed = response.status === 200;
    const blocked = response.status === 429;

    /*
     * Adding zero ensures the unexpected_responses metric
     * exists even when no unexpected response occurs.
     */
    unexpectedResponses.add(0);

    if (allowed) {
        allowedResponses.add(1);
    } else if (blocked) {
        blockedResponses.add(1);
    } else {
        unexpectedResponses.add(1);
    }

    check(response, {
        "status is 200 or 429": () =>
            allowed || blocked,

        "allowed response is valid": () =>
            !allowed ||
            (
                Array.isArray(body.results) &&
                typeof body.count === "number"
            ),

        "blocked response contains rate-limit error": () =>
            !blocked ||
            body.error === "rate limit exceeded",
    });
}