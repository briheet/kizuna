import { expect, it, describe } from "vitest";

const baseUrl = "http://localhost:4000/api/v1";

describe("reach health handler", () => {
  it("returns 200", async () => {
    const response = await fetch(`${baseUrl}/health`, {
      method: "GET",
      headers: {
        "content-type": "application/json",
      },
    });

    expect(response.status).toBe(200);
    expect(response.headers.get("content-type")).toContain("application/json");
  });
});
