import { test, expect } from "../../support/multi-tenant";
import { scenarioTC0237 } from "../../support/multi-tenant-scenarios";

test.describe("TC-237 auto-enabled monitor menus", () => {
  test.use({ multiTenantMode: "multi-tenant-enabled" });

  test("TC-237a: tenant routes include auto-enabled monitor plugins", async ({
    multiTenantMode,
  }) => {
    expect(multiTenantMode).toBe("multi-tenant-enabled");
    await scenarioTC0237();
  });
});
