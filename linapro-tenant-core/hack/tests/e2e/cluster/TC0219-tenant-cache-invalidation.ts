import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0219 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-219 租户缓存失效隔离', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-219a: cache revision rows are isolated by tenant scope', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0219();
  });
});
