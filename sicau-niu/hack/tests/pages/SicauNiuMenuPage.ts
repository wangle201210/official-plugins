import { expect, type Page } from "@host-tests/support/playwright";

import { SicauNiuOperatorPage } from "./SicauNiuOperatorPage";

// SicauNiuMenuPage focuses on the plugin sidebar grouping contract. It verifies
// that the operator pages are reachable through three independent top-level
// responsibility directories rather than a third-level nested menu.
export class SicauNiuMenuPage extends SicauNiuOperatorPage {
  constructor(page: Page) {
    super(page);
  }

  async expectResponsibilityGroupsVisible() {
    await this.expectSidebarDirectoryVisible("寻牛配置");
    await this.expectSidebarDirectoryVisible("寻牛运营");
    await this.expectSidebarDirectoryVisible("寻牛记录");
  }

  async expectLegacyRootMenuAbsent() {
    await this.expectSidebarDirectoryAbsent("寻牛活动");
  }

  async expectResponsibilityGroupsOpenIndependently() {
    await this.collapseSidebarDirectory("寻牛配置");
    await this.collapseSidebarDirectory("寻牛运营");
    await this.collapseSidebarDirectory("寻牛记录");

    await this.expandSidebarDirectory("寻牛配置");
    await this.expectSidebarDirectoryOpen("寻牛配置");
    await this.expectSidebarDirectoryClosed("寻牛运营");
    await this.expectSidebarDirectoryClosed("寻牛记录");

    await this.expandSidebarDirectory("寻牛运营");
    await this.expectSidebarDirectoryClosed("寻牛配置");
    await this.expectSidebarDirectoryOpen("寻牛运营");
    await this.expectSidebarDirectoryClosed("寻牛记录");

    await this.expandSidebarDirectory("寻牛记录");
    await this.expectSidebarDirectoryClosed("寻牛配置");
    await this.expectSidebarDirectoryClosed("寻牛运营");
    await this.expectSidebarDirectoryOpen("寻牛记录");
  }

  async expectGroupedPageReachable(groupName: string, menuName: string) {
    await this.expandSidebarDirectory(groupName);
    await expect(this.sidebarMenuItem(menuName)).toBeVisible();
    await this.clickSidebarMenuItem(menuName);
  }
}
