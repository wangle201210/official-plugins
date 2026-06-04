import { expect, type Locator, type Page } from "@host-tests/support/playwright";

import { PluginPage } from "@host-tests/pages/PluginPage";

// SicauNiuSettlementPage drives the sicau-niu operator settlement page owned by the
// C7 niu-settlement change and extended by C8 (activity), C9/C7 and C10 (anomaly).
// It exposes sidebar navigation plus the dashboard / activity / risk / anomaly /
// archive sections and a create-archive flow whose assertion anchors on the
// persisted archive table row rather than only a button click, so the E2E proves
// settlement state. The "寻牛活动" parent directory is auto-expanded by
// clickSidebarMenuItem, so the nested "运营结算" item is clickable directly.
export class SicauNiuSettlementPage extends PluginPage {
  constructor(page: Page) {
    super(page);
  }

  dashboard(): Locator {
    return this.page.getByTestId("settlement-dashboard").first();
  }

  activitySection(): Locator {
    return this.page.getByTestId("settlement-activity").first();
  }

  riskTable(): Locator {
    return this.page.getByTestId("settlement-risk-table").first();
  }

  anomalyTable(): Locator {
    return this.page.getByTestId("settlement-anomaly-table").first();
  }

  archiveTitleInput(): Locator {
    return this.page.getByTestId("settlement-archive-title").last();
  }

  archiveButton(): Locator {
    return this.page.getByTestId("settlement-archive").first();
  }

  archiveTable(): Locator {
    return this.page.getByTestId("settlement-archive-table").first();
  }

  // openSettlementFromMenu navigates to the operator settlement page and waits for
  // the dashboard section to render, proving the page mounted and loaded.
  async openSettlementFromMenu() {
    await this.clickSidebarMenuItem("运营结算");
    await expect(this.dashboard()).toBeVisible();
  }

  // createArchive fills the archive title, creates the snapshot and asserts the new
  // archive row appears in the archive table.
  async createArchive(title: string) {
    await this.archiveTitleInput().fill(title);
    await this.archiveButton().click();
    await expect(this.archiveTable().locator(".ant-table-row", { hasText: title }).first()).toBeVisible();
  }
}
