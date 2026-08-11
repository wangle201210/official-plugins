import { expect, type Locator, type Page } from "@host-tests/support/playwright";

import { SicauNiuOperatorPage } from "./SicauNiuOperatorPage";

// SicauNiuCatalogPage drives the sicau-niu operator content-asset pages owned by
// the C2 niu-catalog-admin change: the "金句管理" (quotes) and "牛管理" (niu) lists.
// It exposes sidebar navigation plus create / edit / delete flows whose
// assertions anchor on the page data-testid attributes and the persisted vxe row
// text, so the E2E proves list state rather than only button clicks. The
// Asset pages now live under the "寻牛配置" directory, so navigation expands the
// group before clicking the target page.
export class SicauNiuCatalogPage extends SicauNiuOperatorPage {
  constructor(page: Page) {
    super(page);
  }

  // ---------------------------------------------------------------------------
  // Shared helpers
  // ---------------------------------------------------------------------------

  private rowByText(text: string): Locator {
    return this.page.locator(".vxe-body--row", { hasText: text }).first();
  }

  // selectAntOption opens an ant-design Select identified by its form
  // data-testid wrapper and picks the option whose text matches exactly. The
  // testid is rendered on the .ant-select wrapper, so clicking it toggles the
  // dropdown; the option is then chosen from the visible dropdown overlay.
  private async selectAntOption(testId: string, optionLabel: string) {
    await this.page.getByTestId(testId).first().click();
    const dropdown = this.page.locator(".ant-select-dropdown:visible").last();
    await expect(dropdown).toBeVisible();
    await dropdown
      .locator(".ant-select-item-option", { hasText: optionLabel })
      .filter({ hasText: optionLabel })
      .first()
      .click();
    await expect(dropdown).toBeHidden();
  }

  private async confirmDialog(dialog: Locator) {
    await dialog
      .getByRole("button", { name: /确\s*认|确\s*定/i })
      .last()
      .click();
    await expect(dialog).toHaveCount(0);
  }

  // confirmRowPopconfirm clicks the row action that triggers an ant Popconfirm
  // and confirms the popover, mirroring the C1 delete flow.
  private async confirmPopconfirm() {
    const confirmPopover = this.page.locator(".ant-popover:visible").last();
    await expect(confirmPopover).toBeVisible();
    await confirmPopover
      .getByRole("button", { name: /确\s*定|确\s*认/i })
      .click();
  }

  // ---------------------------------------------------------------------------
  // 金句管理 (quotes) — simplest asset (content + enabled), the stable core path
  // ---------------------------------------------------------------------------

  quoteAddButton(): Locator {
    return this.page.getByTestId("sicau-niu-quote-add").first();
  }

  quoteModal(): Locator {
    return this.page
      .getByRole("dialog", { name: /新增金句|编辑金句/ })
      .last();
  }

  quoteContentInput(): Locator {
    return this.page.getByTestId("sicau-niu-quote-content-input").last();
  }

  quoteRow(content: string): Locator {
    return this.rowByText(content);
  }

  async openQuoteFromMenu() {
    await this.openGroupedMenu("寻牛配置", "金句管理");
    await expect(this.quoteAddButton()).toBeVisible();
  }

  async createQuote(content: string) {
    await expect(this.quoteAddButton()).toBeVisible();
    await this.quoteAddButton().click();
    await expect(this.quoteModal()).toBeVisible();
    await this.quoteContentInput().fill(content);
    await this.confirmDialog(this.quoteModal());
    await expect(this.quoteRow(content)).toBeVisible();
  }

  async editQuote(currentContent: string, nextContent: string) {
    const editButton = await this.rowActionButton(currentContent, /编\s*辑/);
    await editButton.click();
    await expect(this.quoteModal()).toBeVisible();
    await expect(this.quoteContentInput()).toHaveValue(currentContent);
    await this.quoteContentInput().fill(nextContent);
    await this.confirmDialog(this.quoteModal());
    await expect(this.quoteRow(nextContent)).toBeVisible();
  }

  async deleteQuote(content: string) {
    const deleteButton = await this.rowActionButton(content, /删\s*除/);
    await deleteButton.click();
    await this.confirmPopconfirm();
    await expect(this.quoteRow(content)).toHaveCount(0);
  }

  // ---------------------------------------------------------------------------
  // 牛管理 (niu) — common niu only, avoiding special-subtype / college coupling
  // ---------------------------------------------------------------------------

  niuAddButton(): Locator {
    return this.page.getByTestId("sicau-niu-niu-add").first();
  }

  niuModal(): Locator {
    return this.page.getByRole("dialog", { name: /新增牛|编辑牛/ }).last();
  }

  niuCodeInput(): Locator {
    return this.page.getByTestId("sicau-niu-niu-code-input").last();
  }

  niuLatInput(): Locator {
    return this.page
      .getByTestId("sicau-niu-niu-lat-input")
      .locator("input")
      .last();
  }

  niuLngInput(): Locator {
    return this.page
      .getByTestId("sicau-niu-niu-lng-input")
      .locator("input")
      .last();
  }

  niuRow(code: string): Locator {
    return this.rowByText(code);
  }

  async openNiuFromMenu() {
    await this.openGroupedMenu("寻牛配置", "牛管理");
    await expect(this.niuAddButton()).toBeVisible();
  }

  // createCommonNiu creates a "普通" niu with just its code. GPS lat/lng default
  // to 0 server-side, so the brittle InputNumber widgets are avoided; the only
  // required field beyond code is niuType, which we set via its Select. The
  // common type keeps the special-subtype / college / name fields hidden.
  async createCommonNiu(code: string) {
    await expect(this.niuAddButton()).toBeVisible();
    await this.niuAddButton().click();
    await expect(this.niuModal()).toBeVisible();
    await this.niuCodeInput().fill(code);
    // niuType defaults to "common" and the form submits the default value, so we
    // fill only the code and submit — no Select/InputNumber interaction, which
    // keeps the create flow free of dropdown-overlay flakiness.
    await this.confirmDialog(this.niuModal());
    await expect(this.niuRow(code)).toBeVisible();
  }

  async editNiuCode(currentCode: string, nextCode: string) {
    const editButton = await this.rowActionButton(currentCode, /编\s*辑/);
    await editButton.click();
    await expect(this.niuModal()).toBeVisible();
    await expect(this.niuCodeInput()).toHaveValue(currentCode);
    await this.niuCodeInput().fill(nextCode);
    await this.confirmDialog(this.niuModal());
    await expect(this.niuRow(nextCode)).toBeVisible();
  }

  async deleteNiu(code: string) {
    const deleteButton = await this.rowActionButton(code, /删\s*除/);
    await deleteButton.click();
    await this.confirmPopconfirm();
    await expect(this.niuRow(code)).toHaveCount(0);
  }

  // ---------------------------------------------------------------------------
  // 铁牛管理 (iron) — location is read-only; create can update the IOT
  // reporting cycle.
  // ---------------------------------------------------------------------------

  ironAddButton(): Locator {
    return this.page.getByTestId("sicau-niu-iron-add").first();
  }

  ironModal(): Locator {
    return this.page.getByRole("dialog", { name: /新增铁牛|编辑铁牛/ }).last();
  }

  ironCodeInput(): Locator {
    return this.page.getByTestId("sicau-niu-iron-code-input").last();
  }

  ironReportingCycleButton(): Locator {
    return this.page.getByTestId("sicau-niu-iron-reporting-cycle").last();
  }

  async openIronFromMenu() {
    await this.openGroupedMenu("寻牛配置", "铁牛管理");
    await expect(this.ironAddButton()).toBeVisible();
  }

  async openNewIronModal() {
    await this.ironAddButton().click();
    await expect(this.ironModal()).toBeVisible();
    await expect(this.ironReportingCycleButton()).toBeVisible();
  }

  async updateIronReportingCycle(code: string) {
    await this.ironCodeInput().fill(code);
    await this.ironReportingCycleButton().click();
  }

  async expectIronLocationReadonly() {
    await expect(this.tableColumn("纬度")).toBeVisible();
    await expect(this.tableColumn("经度")).toBeVisible();
    await expect(this.tableColumn("最近同步时间")).toBeVisible();

    const rowWithCoordinate = this.page
      .locator(".vxe-body--row")
      .filter({ hasText: /\d{2}\.\d{6}/ })
      .first();
    await expect(rowWithCoordinate).toBeVisible();
    await expect(rowWithCoordinate).toContainText(/\d{2}\.\d{6}/);

    await this.page.getByRole("button", { name: /编\s*辑/ }).first().click();
    const snapshot = this.ironModal().getByTestId(
      "sicau-niu-iron-location-snapshot",
    );
    await expect(snapshot).toBeVisible();
    await expect(snapshot).toContainText("定位信息");
    await expect(snapshot).toContainText("纬度");
    await expect(snapshot).toContainText("经度");
    await expect(snapshot.locator("input, textarea")).toHaveCount(0);
    await this.ironModal().getByRole("button", { name: /取\s*消/ }).click();
  }

  // rowActionButton locates a row by its text and returns the named action
  // button (编辑 / 删除) within that row, matching the C1 page-object pattern.
  private async rowActionButton(text: string, action: RegExp) {
    const row = this.rowByText(text);
    await expect(row, `未找到列表行: ${text}`).toBeVisible();
    return row.getByRole("button", { name: action }).first();
  }
}
