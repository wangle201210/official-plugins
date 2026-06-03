import { test, expect } from "@host-tests/fixtures/auth";
import { ensureSourcePluginEnabled } from "@host-tests/fixtures/plugin";

import { SicauNiuHonorPage } from "../pages/SicauNiuHonorPage";

const pluginID = "sicau-niu";

// TC-3 covers the sicau-niu operator honor-config CRUD, the primary
// user-observable path of the C5 niu-ranking-honor change's operator page. It
// creates a honor definition and asserts it persists in the list, edits its name
// and asserts the updated row, then deletes it and asserts removal. A unique
// run-scoped suffix keeps each test self-contained and order-independent.
//
// The create flow only fills 编码 (code) + 名称 (name); the 类型 and 解锁规则
// Selects submit their form defaultValue (徽章 / 参与即得), and 阈值/排序 default
// to 0, so no dropdown overlay is clicked on create — keeping the path stable
// under Playwright. The 集齐分类 category Select stays hidden for the default
// participation rule.
test.describe("TC-3 sicau-niu 荣誉配置 CRUD", () => {
  let honorPage: SicauNiuHonorPage;
  const suffix = `${Date.now()}`;

  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, pluginID);
    honorPage = new SicauNiuHonorPage(adminPage);
  });

  test("TC-3a: 新增荣誉后列表出现该记录", async () => {
    const code = `HONOR-${suffix}-a`;
    const name = `川农荣誉-${suffix}-a`;
    await honorPage.openHonorFromMenu();
    await honorPage.createHonor(code, name);
    await expect(honorPage.honorRow(code)).toBeVisible();
  });

  test("TC-3b: 编辑荣誉名称后列表更新", async () => {
    const code = `HONOR-${suffix}-b`;
    const name = `川农荣誉-${suffix}-b`;
    const renamed = `川农荣誉-改-${suffix}-b`;
    await honorPage.openHonorFromMenu();
    await honorPage.createHonor(code, name);
    await honorPage.editHonorName(code, renamed);
    await expect(honorPage.honorRow(renamed)).toBeVisible();
  });

  test("TC-3c: 删除荣誉后列表移除该记录", async () => {
    const code = `HONOR-${suffix}-c`;
    const name = `川农荣誉-待删-${suffix}-c`;
    await honorPage.openHonorFromMenu();
    await honorPage.createHonor(code, name);
    await honorPage.deleteHonor(code);
    await expect(honorPage.honorRow(code)).toHaveCount(0);
  });
});
