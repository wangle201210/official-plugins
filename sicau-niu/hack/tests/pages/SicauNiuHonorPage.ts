import { expect, type Locator, type Page } from "@host-tests/support/playwright";

import { SicauNiuOperatorPage } from "./SicauNiuOperatorPage";

// SicauNiuHonorPage drives the sicau-niu operator honor-config page owned by the
// C5 niu-ranking-honor change: the "荣誉配置" list. It exposes sidebar navigation
// plus create / edit / delete flows whose assertions anchor on the page
// data-testid attributes and the persisted vxe row text, so the E2E proves list
// state rather than only button clicks. The page lives under the "寻牛配置"
// directory, so navigation expands that group before clicking it. The honor
// form's 类型 / 解锁规则 Selects are NOT interacted with on
// create; the form submits their defaultValue (badge / participation), which
// keeps the create flow free of dropdown-overlay flakiness.
export class SicauNiuHonorPage extends SicauNiuOperatorPage {
  constructor(page: Page) {
    super(page);
  }

  private rowByText(text: string): Locator {
    return this.page.locator(".vxe-body--row", { hasText: text }).first();
  }

  private async confirmDialog(dialog: Locator) {
    await dialog
      .getByRole("button", { name: /确\s*认|确\s*定/i })
      .last()
      .click();
    await expect(dialog).toHaveCount(0);
  }

  private async confirmPopconfirm() {
    const confirmPopover = this.page.locator(".ant-popover:visible").last();
    await expect(confirmPopover).toBeVisible();
    await confirmPopover
      .getByRole("button", { name: /确\s*定|确\s*认/i })
      .click();
  }

  honorAddButton(): Locator {
    return this.page.getByTestId("sicau-niu-honor-add").first();
  }

  honorModal(): Locator {
    return this.page
      .getByRole("dialog", { name: /新增荣誉|编辑荣誉/ })
      .last();
  }

  honorCodeInput(): Locator {
    return this.page.getByTestId("sicau-niu-honor-code-input").last();
  }

  honorNameInput(): Locator {
    return this.page.getByTestId("sicau-niu-honor-name-input").last();
  }

  honorRow(text: string): Locator {
    return this.rowByText(text);
  }

  async openHonorFromMenu() {
    await this.openGroupedMenu("寻牛配置", "荣誉配置");
    await expect(this.honorAddButton()).toBeVisible();
  }

  // createHonor creates a honor with just its code + name. The 类型 (defaults to
  // "徽章"/badge) and 解锁规则 (defaults to "参与即得"/participation) Selects are
  // submitted via their form defaultValue, so the brittle dropdown overlays are
  // not clicked; 阈值/排序 default to 0 and 分类 stays hidden for participation.
  async createHonor(code: string, name: string) {
    await expect(this.honorAddButton()).toBeVisible();
    await this.honorAddButton().click();
    await expect(this.honorModal()).toBeVisible();
    await this.honorCodeInput().fill(code);
    await this.honorNameInput().fill(name);
    await this.confirmDialog(this.honorModal());
    await expect(this.honorRow(code)).toBeVisible();
  }

  async editHonorName(code: string, nextName: string) {
    const editButton = await this.rowActionButton(code, /编\s*辑/);
    await editButton.click();
    await expect(this.honorModal()).toBeVisible();
    await expect(this.honorCodeInput()).toHaveValue(code);
    await this.honorNameInput().fill(nextName);
    await this.confirmDialog(this.honorModal());
    await expect(this.honorRow(nextName)).toBeVisible();
  }

  async deleteHonor(code: string) {
    const deleteButton = await this.rowActionButton(code, /删\s*除/);
    await deleteButton.click();
    await this.confirmPopconfirm();
    await expect(this.honorRow(code)).toHaveCount(0);
  }

  // rowActionButton locates a row by its text and returns the named action
  // button (编辑 / 删除) within that row, matching the C2 page-object pattern.
  private async rowActionButton(text: string, action: RegExp) {
    const row = this.rowByText(text);
    await expect(row, `未找到列表行: ${text}`).toBeVisible();
    return row.getByRole("button", { name: action }).first();
  }
}
