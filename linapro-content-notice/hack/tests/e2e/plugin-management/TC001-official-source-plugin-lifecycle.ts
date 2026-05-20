import type { Page } from '@host-tests/support/playwright';

import { test, expect } from '@host-tests/fixtures/auth';
import {
  createAdminApiContext,
  ensureSourcePluginEnabled,
  ensureSourcePluginEnabledViaAPI,
  ensureSourcePluginUninstalled,
  findPlugin,
  installPlugin,
  refreshPluginProjection,
  syncPlugins,
  uninstallPlugin,
  updatePluginStatus,
} from '@host-tests/fixtures/plugin';
import {
  expectMountedTitles,
  expectPluginRouteAvailable,
  expectPluginRouteMissing,
  expectPluginState,
  expectRuntimePluginState,
  type SourcePluginLifecycleCase,
} from '@host-tests/support/source-plugin-lifecycle';

const contentNoticePluginCase: SourcePluginLifecycleCase = {
  id: 'linapro-content-notice',
  mountedTitles: ['通知公告'],
  route: '/system/notice',
  assertAvailable: async (page: Page) => {
    await expect(page.locator('.vxe-table')).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('button', { name: /新\s*增/ }).first()).toBeVisible();
  },
};

test.describe('TC-1 官方源码插件生命周期', () => {
  test(`TC001a: ${contentNoticePluginCase.id} 支持完整安装、启用、停用、卸载与菜单挂载切换`, async ({
    adminPage,
  }) => {
    test.setTimeout(180_000);
    const adminApi = await createAdminApiContext();
    const item = contentNoticePluginCase;

    try {
      await syncPlugins(adminApi);

      await ensureSourcePluginUninstalled(adminPage, item.id);
      await expectPluginState(adminApi, item.id, 0, 0);
      await expectRuntimePluginState(adminApi, item.id, 0, 0);
      await refreshPluginProjection(adminPage);
      await expectMountedTitles(adminApi, item.mountedTitles, false);
      await expectPluginRouteMissing(adminPage, item.route);

      await installPlugin(adminApi, item.id);
      await expectPluginState(adminApi, item.id, 1, 0);
      await expectRuntimePluginState(adminApi, item.id, 1, 0);
      await refreshPluginProjection(adminPage);
      await expectMountedTitles(adminApi, item.mountedTitles, false);
      await expectPluginRouteMissing(adminPage, item.route);

      await updatePluginStatus(adminApi, item.id, true);
      await expectPluginState(adminApi, item.id, 1, 1);
      await expectRuntimePluginState(adminApi, item.id, 1, 1);
      await refreshPluginProjection(adminPage);
      await expectMountedTitles(adminApi, item.mountedTitles, true);
      await expectPluginRouteAvailable(adminPage, item);

      await updatePluginStatus(adminApi, item.id, false);
      await expectPluginState(adminApi, item.id, 1, 0);
      await expectRuntimePluginState(adminApi, item.id, 1, 0);
      await refreshPluginProjection(adminPage);
      await expectMountedTitles(adminApi, item.mountedTitles, false);
      await expectPluginRouteMissing(adminPage, item.route);

      await uninstallPlugin(adminApi, item.id);
      await expectPluginState(adminApi, item.id, 0, 0);
      await expectRuntimePluginState(adminApi, item.id, 0, 0);
      await refreshPluginProjection(adminPage);
      await expectMountedTitles(adminApi, item.mountedTitles, false);
      await expectPluginRouteMissing(adminPage, item.route);
    } finally {
      try {
        await ensureSourcePluginEnabledViaAPI(adminApi, item.id);
      } finally {
        await adminApi.dispose();
      }
      if (!adminPage.isClosed()) {
        await ensureSourcePluginEnabled(adminPage, item.id);
      }
    }
  });
});
