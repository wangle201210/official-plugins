import { test, expect } from "@host-tests/fixtures/auth";
import { ensureSourcePluginEnabled } from "@host-tests/fixtures/plugin";

const pluginID = "sicau-niu";

// The host serves the C6 H5 memorial wall as a public static asset under
// /x-assets/{plugin-id}/{version}/. The version segment matches plugin.yaml.
const wallPath = "/x-assets/sicau-niu/v0.2.0/wall/index.html";
const publicApiBase = "/x/sicau-niu/api/v1/plugins/sicau-niu/wall";

// TC-5 closes the C6 niu-h5-wall public-page acceptance: with NO login it visits
// the host-served H5 memorial wall and confirms the page renders and its public
// (no-auth) API is reachable, and it independently asserts the public API returns
// a success envelope without any token. The plugin is ensured enabled via the
// admin context in beforeEach, but the assertions use the anonymous `page` fixture
// (no admin storage state) so the test genuinely proves no-token access.
test.describe("TC-5 sicau-niu H5 公开纪念墙", () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, pluginID);
  });

  test("TC-5a: 无登录访问 H5 纪念墙页且公开数据渲染", async ({ page }) => {
    await page.goto(wallPath, { waitUntil: "domcontentloaded" });

    // Static structure renders.
    await expect(page.locator("h1")).toContainText("数字纪念墙");
    await expect(page.getByText("首发纪念墙")).toBeVisible();

    // The public stats API was fetched and rendered: the total-cattle stat leaves
    // its "—" placeholder once the no-auth API call resolves.
    await expect(page.locator('[data-k="totalNiuCount"]')).not.toHaveText("—", {
      timeout: 15000,
    });
  });

  test("TC-5b: 无 token 公开 API 返回成功包络", async ({ page }) => {
    const res = await page.request.get(`${publicApiBase}/stats`);
    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(body.code).toBe(0);
    expect(body.data).toBeTruthy();

    const config = await page.request.get(`${publicApiBase}/config`);
    expect(config.status()).toBe(200);
    expect((await config.json()).code).toBe(0);
  });
});
