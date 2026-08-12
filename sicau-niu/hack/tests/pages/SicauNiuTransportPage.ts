import {
  expect,
  type Locator,
  type Page,
} from "@host-tests/support/playwright";

import { SicauNiuOperatorPage } from "./SicauNiuOperatorPage";

// SicauNiuTransportPage drives the operator cloud-moving team and report audit
// workflow. Selectors are scoped to page-specific test ids and VXE rows so the
// assertions prove persisted API results rather than only menu navigation.
export class SicauNiuTransportPage extends SicauNiuOperatorPage {
  constructor(page: Page) {
    super(page);
  }

  stats(): Locator {
    return this.page.getByTestId("sicau-niu-transport-stats").first();
  }

  teamRow(name: string): Locator {
    return this.page.locator(".vxe-body--row", { hasText: name }).first();
  }

  renameDialog(): Locator {
    return this.page.getByRole("dialog", { name: "修改团名称" }).last();
  }

  renameInput(): Locator {
    return this.page.getByTestId("sicau-niu-transport-name-input").last();
  }

  async openFromMenu() {
    await this.openGroupedMenu("寻牛运营", "云搬牛管理");
    await expect(this.stats()).toBeVisible();
  }

  async expectSeededTeam(name: string, contributionMeters: number) {
    const row = this.teamRow(name);
    await expect(row).toBeVisible();
    await expect(row).toContainText(
      `${contributionMeters.toLocaleString("zh-CN")} 米`,
    );
  }

  async renameTeam(currentName: string, nextName: string) {
    const row = this.teamRow(currentName);
    await expect(row).toBeVisible();
    await row.getByRole("button", { name: "改名" }).click();
    await expect(this.renameDialog()).toBeVisible();
    await expect(this.renameInput()).toHaveValue(currentName);
    await this.renameInput().fill(nextName);
    await this.renameDialog()
      .getByRole("button", { name: /确\s*认|确\s*定/u })
      .last()
      .click();
    await expect(this.renameDialog()).toHaveCount(0);
    await expect(this.teamRow(nextName)).toBeVisible();
  }

  async expectDuplicateRenameRejected(currentName: string, duplicateName: string) {
    const row = this.teamRow(currentName);
    await expect(row).toBeVisible();
    await row.getByRole("button", { name: "改名" }).click();
    await expect(this.renameDialog()).toBeVisible();
    await this.renameInput().fill(duplicateName);
    await this.renameDialog()
      .getByRole("button", { name: /确\s*认|确\s*定/u })
      .last()
      .click();
    await expect(this.page.getByText("这个有效团名已经被使用").last()).toBeVisible();
    await expect(this.renameDialog()).toBeVisible();
    await expect(this.renameInput()).toHaveValue(duplicateName);
  }

  async openReportAudit() {
    await this.page.getByRole("tab", { name: "位置上报审计" }).click();
    await expect(
      this.page.locator(".vxe-header--column", { hasText: "本次贡献" }).first(),
    ).toBeVisible();
  }

  async expectReportFacts(teamName: string, contributionMeters: number) {
    const rows = this.page.locator(".vxe-body--row", { hasText: teamName });
    await expect(rows).toHaveCount(2);
    await expect(
      rows.filter({
        hasText: `${contributionMeters.toLocaleString("zh-CN")} 米`,
      }),
    ).toHaveCount(1);
    await expect(rows.filter({ hasText: "首次上报" })).toHaveCount(1);
  }
}
