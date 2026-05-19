import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0192 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-192 通知跨租户隔离', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-192a: tenant notices and platform broadcast messages persist separately', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0192();
  });
});
