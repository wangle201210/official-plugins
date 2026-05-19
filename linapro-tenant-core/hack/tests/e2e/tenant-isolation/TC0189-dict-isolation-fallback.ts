import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0189 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-189 字典跨租户隔离', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-189a: tenant dictionary override and platform fallback rows stay isolated', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0189();
  });
});
