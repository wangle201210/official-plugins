import { test, expect } from "@host-tests/fixtures/auth";
import { ensureSourcePluginEnabled } from "@host-tests/fixtures/plugin";

import { SicauNiuSettlementPage } from "../pages/SicauNiuSettlementPage";

const pluginID = "sicau-niu";

// TC-4 covers the sicau-niu operator settlement page, the primary user-observable
// path of the C7 niu-settlement change plus its C8 activity and C10 anomaly
// extensions. It asserts the page mounts with its dashboard, activity, risk and
// anomaly sections rendered, then creates a settlement archive and asserts the
// snapshot persists in the archive table. A unique run-scoped suffix keeps each
// test self-contained and order-independent. The archive flow only fills the title
// input and clicks the create button — no dropdown overlay is involved — keeping
// the path stable under Playwright.
test.describe("TC-4 sicau-niu 运营结算", () => {
  let settlementPage: SicauNiuSettlementPage;
  const suffix = `${Date.now()}`;

  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, pluginID);
    settlementPage = new SicauNiuSettlementPage(adminPage);
  });

  test("TC-4a: 运营结算页加载且各区块可见", async () => {
    await settlementPage.openSettlementFromMenu();
    await expect(settlementPage.dashboard()).toBeVisible();
    await expect(settlementPage.overviewTitle()).toBeVisible();
    await expect(settlementPage.activitySection()).toBeVisible();
    await expect(settlementPage.rulesSection()).toBeVisible();
    await expect(settlementPage.rankingsSection()).toBeVisible();
    await expect(settlementPage.keyInteractions()).toBeVisible();
    await expect(settlementPage.archiveActionDescription()).toBeVisible();
    const archiveDescriptionBox = await settlementPage
      .archiveActionDescription()
      .boundingBox();
    expect(archiveDescriptionBox?.width ?? 0).toBeGreaterThan(180);
    await expect(settlementPage.riskTable()).toBeVisible();
    await expect(settlementPage.anomalyTable()).toBeVisible();
    await expect(settlementPage.archiveTable()).toBeVisible();
    await expect(settlementPage.rulesSection()).toContainText("互动规则配置");
    await expect(settlementPage.rankingsSection()).toContainText("排行榜数据");
  });

  test("TC-4b: 创建结算归档后归档表出现该记录", async () => {
    const title = `结算公示-${suffix}`;
    await settlementPage.openSettlementFromMenu();
    await settlementPage.createArchive(title);
    await expect(
      settlementPage
        .archiveTable()
        .locator(".ant-table-row", { hasText: title })
        .first(),
    ).toBeVisible();
  });

  test("TC-4c: 选择证书后可以执行批量发证", async () => {
    await settlementPage.openSettlementFromMenu();
    await settlementPage.issueFirstCertificate();
  });
});
