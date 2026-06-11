import { test, expect } from "@host-tests/fixtures/auth";
import { ensureSourcePluginEnabled } from "@host-tests/fixtures/plugin";

import { SicauNiuMenuPage } from "../pages/SicauNiuMenuPage";

const pluginID = "sicau-niu";

// TC-7 covers the sicau-niu operator menu grouping feedback. It asserts three
// independent top-level responsibility directories and verifies each directory
// can still open a representative page without changing routes or permissions.
test.describe("TC-7 sicau-niu 菜单拆分", () => {
  let menuPage: SicauNiuMenuPage;

  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, pluginID);
    menuPage = new SicauNiuMenuPage(adminPage);
  });

  test("TC-7a: 寻牛菜单按职责拆成独立一级入口", async () => {
    await menuPage.expectResponsibilityGroupsVisible();
    await menuPage.expectLegacyRootMenuAbsent();
  });

  test("TC-7b: 各一级入口可单独控制开合状态", async () => {
    await menuPage.expectResponsibilityGroupsOpenIndependently();
  });

  test("TC-7c: 各一级入口下代表页面保持可进入", async () => {
    await menuPage.expectGroupedPageReachable("寻牛配置", "牛管理");
    await expect(menuPage.page).toHaveURL(/sicau-niu-niu/);

    await menuPage.expectGroupedPageReachable("寻牛运营", "运营结算");
    await expect(menuPage.page).toHaveURL(/sicau-niu-settlement/);

    await menuPage.expectGroupedPageReachable("寻牛记录", "激活记录");
    await expect(menuPage.page).toHaveURL(/sicau-niu-record-activation/);
  });
});
