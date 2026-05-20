import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0194 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-4 在线会话跨租户隔离', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-4a: online session revocation targets tenant-token pairs', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0194();
  });
});
