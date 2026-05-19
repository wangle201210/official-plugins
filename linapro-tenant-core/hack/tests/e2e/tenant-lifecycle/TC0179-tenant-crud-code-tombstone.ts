import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0179 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-179 平台管理员创建租户', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-179a: tenant CRUD validates code uniqueness and tombstones', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0179();
  });
});
