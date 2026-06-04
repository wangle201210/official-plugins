import { test, expect } from "@host-tests/fixtures/auth";
import { ensureSourcePluginEnabled } from "@host-tests/fixtures/plugin";

import { SicauNiuCollegePage } from "../pages/SicauNiuCollegePage";

const pluginID = "sicau-niu";

// TC-1 covers the sicau-niu operator college dictionary CRUD, the primary
// user-observable path of the C1 niu-identity change: create a college and
// assert it persists in the list, rename it and assert the update, reject a
// duplicate name, then delete it and assert removal. A unique run-scoped suffix
// keeps the test self-contained and order-independent.
test.describe("TC-1 sicau-niu 院系字典 CRUD", () => {
  let collegePage: SicauNiuCollegePage;
  const suffix = `${Date.now()}`;
  const collegeName = `信息工程学院-${suffix}`;
  const renamedName = `信息工程学院-改-${suffix}`;

  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, pluginID);
    collegePage = new SicauNiuCollegePage(adminPage);
    await collegePage.openFromMenu();
  });

  test("TC-1a: 新增院系后列表出现该记录", async () => {
    await collegePage.createCollege(collegeName, 1);
    await expect(collegePage.collegeRow(collegeName)).toBeVisible();
  });

  test("TC-1b: 编辑院系名称后列表更新", async () => {
    await collegePage.createCollege(`待改-${suffix}`, 2);
    await collegePage.editCollege(`待改-${suffix}`, renamedName);
    await expect(collegePage.collegeRow(renamedName)).toBeVisible();
  });

  test("TC-1c: 同名院系新增被拒绝且不产生重复行", async () => {
    const dupName = `重复学院-${suffix}`;
    await collegePage.createCollege(dupName, 3);
    await collegePage.expectDuplicateNameRejected(dupName);
  });

  test("TC-1d: 删除院系后列表移除该记录", async () => {
    const tempName = `待删-${suffix}`;
    await collegePage.createCollege(tempName, 4);
    await collegePage.deleteCollege(tempName);
    await expect(collegePage.collegeRow(tempName)).toHaveCount(0);
  });
});
