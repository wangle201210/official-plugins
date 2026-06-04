import { test, expect } from "@host-tests/fixtures/auth";
import { ensureSourcePluginEnabled } from "@host-tests/fixtures/plugin";

import { SicauNiuRecordPage } from "../pages/SicauNiuRecordPage";

const pluginID = "sicau-niu";

// TC-6 covers the sicau-niu operator activity-record query pages owned by the
// niu-activity-records change: the read-only feeding, steal and grass-ledger lists
// nested under the "活动记录" sub-directory. It navigates two levels deep and asserts
// each page mounts and its read-only grid renders (a page-unique column header
// becomes visible), proving the page, route and list API are wired. Row presence
// depends on activity data, so the assertions anchor on the grid structure rather
// than a specific row.
test.describe("TC-6 sicau-niu 活动记录查询", () => {
  let recordPage: SicauNiuRecordPage;

  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, pluginID);
    recordPage = new SicauNiuRecordPage(adminPage);
  });

  test("TC-6a: 喂草记录页加载且表格渲染", async () => {
    await recordPage.openRecord("喂草记录");
    await recordPage.expectGridRendered("实际效果");
  });

  test("TC-6b: 偷草记录页加载且表格渲染", async () => {
    await recordPage.openRecord("偷草记录");
    await recordPage.expectGridRendered("被偷玩家");
  });

  test("TC-6c: 草账户流水页加载且表格渲染", async () => {
    await recordPage.openRecord("草账户流水");
    await recordPage.expectGridRendered("增减量");
  });
});
