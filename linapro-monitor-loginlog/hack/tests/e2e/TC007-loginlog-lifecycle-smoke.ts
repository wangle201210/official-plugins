import { test, expect } from '@host-tests/fixtures/auth';
import { smokeSourcePluginLifecycle } from '@host-tests/support/source-plugin-lifecycle';

test.describe('TC007 登录日志插件生命周期 smoke', () => {
  test('TC-3a: 登录日志插件可安装、启用、挂载菜单并访问页面', async ({
    adminPage,
  }) => {
    await smokeSourcePluginLifecycle(adminPage, {
      id: 'linapro-monitor-loginlog',
      mountedTitles: ['登录日志'],
      route: '/monitor/loginlog',
      assertAvailable: async (page) => {
        await expect(page.locator('.vxe-table')).toBeVisible({ timeout: 10000 });
        await expect(page.getByRole('button', { name: /导\s*出/ })).toBeVisible();
      },
    });
  });
});
