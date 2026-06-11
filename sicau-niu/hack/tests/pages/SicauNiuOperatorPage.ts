import { expect, type Page } from "@host-tests/support/playwright";

import { PluginPage } from "@host-tests/pages/PluginPage";

const directoryProbeMenuByName = new Map<string, string>([
  ["寻牛配置", "牛管理"],
  ["寻牛运营", "运营结算"],
  ["寻牛记录", "激活记录"],
]);

export class SicauNiuOperatorPage extends PluginPage {
  constructor(page: Page) {
    super(page);
  }

  private sidebarDirectoryTitle(menuName: string) {
    return this.sidebarMenu
      .locator(".ant-menu-submenu-title, .vben-sub-menu-content")
      .filter({ has: this.page.getByText(menuName, { exact: true }) })
      .first();
  }

  private sidebarDirectory(menuName: string) {
    return this.sidebarDirectoryTitle(menuName)
      .locator(
        "xpath=ancestor::*[contains(concat(' ', normalize-space(@class), ' '), ' ant-menu-submenu ') or contains(concat(' ', normalize-space(@class), ' '), ' vben-sub-menu ')][1]",
      )
      .first();
  }

  private sidebarDirectoryProbeItem(menuName: string) {
    const probeMenuName = directoryProbeMenuByName.get(menuName);
    return probeMenuName ? this.sidebarMenuItem(probeMenuName) : null;
  }

  private async isSidebarDirectoryOpen(menuName: string) {
    const probeItem = this.sidebarDirectoryProbeItem(menuName);
    if (probeItem) {
      return probeItem.isVisible({ timeout: 500 }).catch(() => false);
    }

    const className =
      (await this.sidebarDirectory(menuName).getAttribute("class").catch(() => "")) ??
      "";
    return (
      className.includes("ant-menu-submenu-open") ||
      className.includes("is-opened")
    );
  }

  async expandSidebarDirectory(menuName: string) {
    const title = this.sidebarDirectoryTitle(menuName);
    await expect(title, `未找到侧边栏目录: ${menuName}`).toBeVisible();

    if (!(await this.isSidebarDirectoryOpen(menuName))) {
      await title.click();
    }
  }

  async collapseSidebarDirectory(menuName: string) {
    const title = this.sidebarDirectoryTitle(menuName);
    await expect(title, `未找到侧边栏目录: ${menuName}`).toBeVisible();

    if (await this.isSidebarDirectoryOpen(menuName)) {
      await title.click();
    }
  }

  async expectSidebarDirectoryVisible(menuName: string) {
    await expect(this.sidebarDirectoryTitle(menuName)).toBeVisible();
  }

  async expectSidebarDirectoryAbsent(menuName: string) {
    await expect(this.sidebarDirectoryTitle(menuName)).toHaveCount(0);
  }

  async expectSidebarDirectoryOpen(menuName: string) {
    const probeItem = this.sidebarDirectoryProbeItem(menuName);
    if (probeItem) {
      await expect(probeItem).toBeVisible();
      return;
    }

    await expect(this.sidebarDirectory(menuName)).toHaveClass(
      /(?:ant-menu-submenu-open|is-opened)/,
    );
  }

  async expectSidebarDirectoryClosed(menuName: string) {
    const probeItem = this.sidebarDirectoryProbeItem(menuName);
    if (probeItem) {
      await expect(probeItem).not.toBeVisible();
      return;
    }

    await expect(this.sidebarDirectory(menuName)).not.toHaveClass(
      /(?:ant-menu-submenu-open|is-opened)/,
    );
  }

  async openGroupedMenu(groupName: string, menuName: string) {
    await this.expandSidebarDirectory(groupName);
    await this.clickSidebarMenuItem(menuName);
  }
}
