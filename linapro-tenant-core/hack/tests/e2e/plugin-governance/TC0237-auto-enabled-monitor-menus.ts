import { test, expect } from "../../support/linapro-tenant-core";
import { scenarioTC0237 } from "../../support/linapro-tenant-core-scenarios";

test.describe("TC-237 auto-enabled monitor menus", () => {
  test.use({ multiTenantMode: "linapro-tenant-core-enabled" });

  test("TC-237a: tenant routes include auto-enabled monitor plugins", async ({
    multiTenantMode,
  }) => {
    expect(multiTenantMode).toBe("linapro-tenant-core-enabled");
    await scenarioTC0237();
  });
});
