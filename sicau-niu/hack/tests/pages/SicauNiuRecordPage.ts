import { expect, type Locator, type Page } from "@host-tests/support/playwright";

import { PluginPage } from "@host-tests/pages/PluginPage";

// SicauNiuRecordPage drives the sicau-niu operator activity-record query pages owned
// by the niu-activity-records change: the feeding / steal / gift / check-in /
// activation / grass-ledger read-only lists, mounted directly under the "寻牛活动"
// menu like the other operator pages. Each page is a read-only vxe grid; the
// assertions anchor on a column header unique to the page so the test proves the
// page, route and list API mounted and rendered without depending on any specific
// row being present.
export class SicauNiuRecordPage extends PluginPage {
  constructor(page: Page) {
    super(page);
  }

  // openRecord opens the named record page from the sidebar (the "寻牛活动" parent is
  // auto-expanded by clickSidebarMenuItem).
  async openRecord(menuName: string) {
    await this.clickSidebarMenuItem(menuName);
  }

  // columnHeader returns the grid column header cell with the given title.
  columnHeader(title: string): Locator {
    return this.page.locator(".vxe-header--column", { hasText: title }).first();
  }

  // expectGridRendered asserts the record page's grid rendered by waiting for a
  // column header unique to the page to become visible.
  async expectGridRendered(uniqueHeader: string) {
    await expect(this.columnHeader(uniqueHeader)).toBeVisible();
  }
}
