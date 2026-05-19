import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0196 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-196 部门跨租户隔离', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-196a: tenant department rows use tenant scope without event outbox', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0196();
  });
});
