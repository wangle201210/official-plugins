import { test, expect } from '@host-tests/fixtures/auth';
import { prepareSourcePluginsBaseline } from '@host-tests/fixtures/plugin';
import { NoticePage } from '../../pages/NoticePage';

test.describe('TC001 通知公告 CRUD', () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline(['linapro-content-notice']);
  });

  const testTitle = `测试通知_${Date.now()}`;
  const testTitleRenamed = `${testTitle}_修改`;

  test('TC001a: 创建新通知公告', async ({ adminPage }) => {
    const noticePage = new NoticePage(adminPage);
    await noticePage.goto();
    await noticePage.createNotice(testTitle, '通知', '草稿', '这是测试内容');

    await expect(
      adminPage.getByText(/新增成功|创建成功|success/i),
    ).toBeVisible({ timeout: 5000 });
  });

  test('TC001b: 通知公告列表中可见新创建的记录', async ({ adminPage }) => {
    const noticePage = new NoticePage(adminPage);
    await noticePage.goto();

    const hasNotice = await noticePage.hasNotice(testTitle);
    expect(hasNotice).toBeTruthy();
  });

  test('TC001c: 编辑通知公告', async ({ adminPage }) => {
    const noticePage = new NoticePage(adminPage);
    await noticePage.goto();
    await noticePage.editNotice(testTitle, testTitleRenamed);

    await expect(adminPage.getByText(/更新成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });

  test('TC001d: 删除通知公告', async ({ adminPage }) => {
    const noticePage = new NoticePage(adminPage);
    await noticePage.goto();
    await noticePage.deleteNotice(testTitleRenamed);

    await expect(adminPage.getByText(/删除成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });
});
