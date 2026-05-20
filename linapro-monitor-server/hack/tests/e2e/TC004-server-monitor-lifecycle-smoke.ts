import { test, expect } from '@host-tests/fixtures/auth';
import { smokeSourcePluginLifecycle } from '@host-tests/support/source-plugin-lifecycle';

test.describe('TC004 服务监控插件生命周期 smoke', () => {
  test('TC-4a: 服务监控插件可安装、启用、挂载菜单并访问页面', async ({
    adminPage,
  }) => {
    await smokeSourcePluginLifecycle(adminPage, {
      id: 'linapro-monitor-server',
      mountedTitles: ['服务监控'],
      route: '/monitor/server',
      assertAvailable: async (page) => {
        await expect
          .poll(
            async () => {
              const hasMetrics =
                (await page.getByText('服务信息').first().isVisible()) &&
                (await page.getByText('服务器信息').first().isVisible());
              const hasEmptyState = await page
                .getByText('当前暂无服务监控数据。')
                .first()
                .isVisible();

              return hasMetrics || hasEmptyState ? 'available' : 'pending';
            },
            {
              message: '服务监控页面应展示指标内容或插件空状态',
              timeout: 10000,
            },
          )
          .toBe('available');
      },
    });
  });
});
