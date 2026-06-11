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

  // expectActivationRecordPageHeightStable proves the tab content and active pane
  // do not keep growing after VXE finishes measuring its auto height.
  async expectActivationRecordPageHeightStable() {
    const firstHeight = await this.activationRecordHeightMetrics();
    await this.page.waitForTimeout(800);
    const secondHeight = await this.activationRecordHeightMetrics();
    await this.page.waitForTimeout(800);
    const thirdHeight = await this.activationRecordHeightMetrics();

    for (const key of ["holder", "content", "pane"] as const) {
      const values = [firstHeight[key], secondHeight[key], thirdHeight[key]];
      expect(Math.max(...values) - Math.min(...values)).toBeLessThanOrEqual(4);
    }
  }

  private async activationRecordHeightMetrics() {
    return this.page.evaluate(() => {
      const readScrollHeight = (selector: string) =>
        document.querySelector<HTMLElement>(selector)?.scrollHeight ?? 0;

      return {
        content: readScrollHeight(".activation-record-tabs .ant-tabs-content"),
        holder: readScrollHeight(
          ".activation-record-tabs .ant-tabs-content-holder",
        ),
        pane: readScrollHeight(".activation-record-tabs .ant-tabs-tabpane-active"),
      };
    });
  }

  // activationPhotoButton returns the read-only photo preview trigger for an
  // activation record row.
  activationPhotoButton(id: string): Locator {
    return this.page.getByTestId(`sicau-niu-activation-photo-${id}`).first();
  }

  // activationAttemptPhotoButton returns the photo preview trigger for an
  // activation attempt audit row.
  activationAttemptPhotoButton(id: string): Locator {
    return this.page
      .getByTestId(`sicau-niu-activation-attempt-photo-${id}`)
      .first();
  }

  // expectActivationPhotoPreview opens a record photo and asserts the global image
  // preview layer renders the original image URL.
  async expectActivationPhotoPreview(id: string, photoPath: string) {
    const button = this.activationPhotoButton(id);
    await expect(button).toBeVisible();
    await button.click();
    await expect(
      this.page.locator(".ant-image-preview-img").last(),
    ).toHaveAttribute("src", photoPath);
  }

  // expectActivationAttemptPhotoPreview opens an attempt photo and asserts the
  // global image preview layer renders the original image URL.
  async expectActivationAttemptPhotoPreview(id: string, photoPath: string) {
    const button = this.activationAttemptPhotoButton(id);
    await expect(button).toBeVisible();
    await button.click();
    await expect(
      this.page.locator(".ant-image-preview-img").last(),
    ).toHaveAttribute("src", photoPath);
  }
}
