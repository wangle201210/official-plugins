import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0190 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-190 配置跨租户隔离', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-190a: tenant config override and platform fallback rows stay isolated', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0190();
  });
});
