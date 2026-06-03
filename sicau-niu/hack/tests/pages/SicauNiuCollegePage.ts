import { expect, type Locator, type Page } from "@host-tests/support/playwright";

import { PluginPage } from "@host-tests/pages/PluginPage";

// SicauNiuCollegePage drives the sicau-niu operator "院系字典" (college dictionary)
// page: navigation from the sidebar menu plus create / edit / delete flows used by
// the plugin-owned college CRUD E2E. Selectors anchor on the page data-testid
// attributes and the vxe row text so assertions prove the persisted list state.
export class SicauNiuCollegePage extends PluginPage {
  constructor(page: Page) {
    super(page);
  }

  gridAddButton(): Locator {
    return this.page.getByTestId("sicau-niu-college-add").first();
  }

  collegeModal(): Locator {
    return this.page
      .getByRole("dialog", { name: /新增院系|编辑院系/ })
      .last();
  }

  nameInput(): Locator {
    return this.page.getByTestId("sicau-niu-college-name-input").last();
  }

  sortInput(): Locator {
    return this.page.getByTestId("sicau-niu-college-sort-input").last();
  }

  collegeRow(name: string): Locator {
    return this.page.locator(".vxe-body--row", { hasText: name }).first();
  }

  // openFromMenu navigates to the college dictionary page through the sidebar.
  async openFromMenu() {
    await this.clickSidebarMenuItem("院系字典");
    await expect(this.gridAddButton()).toBeVisible();
  }

  // createCollege opens the modal, fills the form and asserts the new row lands
  // in the persisted list.
  async createCollege(name: string, sort: number) {
    await expect(this.gridAddButton()).toBeVisible();
    await this.gridAddButton().click();
    await expect(this.collegeModal()).toBeVisible();
    await this.nameInput().fill(name);
    await this.sortInput().fill(String(sort));
    await this.confirmModal();
    await expect(this.collegeRow(name)).toBeVisible();
  }

  // editCollege renames an existing college and asserts the updated row.
  async editCollege(currentName: string, nextName: string) {
    const editButton = await this.rowActionButton(currentName, /编\s*辑/);
    await editButton.click();
    await expect(this.collegeModal()).toBeVisible();
    await expect(this.nameInput()).toHaveValue(currentName);
    await this.nameInput().fill(nextName);
    await this.confirmModal();
    await expect(this.collegeRow(nextName)).toBeVisible();
  }

  // expectDuplicateNameRejected asserts that creating a college whose name already
  // exists keeps the modal open and surfaces a business error rather than adding a
  // second row.
  async expectDuplicateNameRejected(name: string) {
    await this.gridAddButton().click();
    await expect(this.collegeModal()).toBeVisible();
    await this.nameInput().fill(name);
    await this.sortInput().fill("0");
    await this.collegeModal()
      .getByRole("button", { name: /确\s*认|确\s*定/i })
      .last()
      .click();
    await expect(this.collegeModal()).toBeVisible();
    await expect(this.page.locator(".vxe-body--row", { hasText: name })).toHaveCount(1);
  }

  // deleteCollege confirms the popconfirm and asserts the row is removed.
  async deleteCollege(name: string) {
    const deleteButton = await this.rowActionButton(name, /删\s*除/);
    await deleteButton.click();
    const confirmPopover = this.page.locator(".ant-popover:visible").last();
    await expect(confirmPopover).toBeVisible();
    await confirmPopover.getByRole("button", { name: /确\s*定|确\s*认/i }).click();
    await expect(this.collegeRow(name)).toHaveCount(0);
  }

  private async confirmModal() {
    await this.collegeModal()
      .getByRole("button", { name: /确\s*认|确\s*定/i })
      .last()
      .click();
    await expect(this.collegeModal()).toHaveCount(0);
  }

  private async rowActionButton(name: string, action: RegExp) {
    const row = this.collegeRow(name);
    await expect(row, `未找到院系行: ${name}`).toBeVisible();
    return row.getByRole("button", { name: action }).first();
  }
}
