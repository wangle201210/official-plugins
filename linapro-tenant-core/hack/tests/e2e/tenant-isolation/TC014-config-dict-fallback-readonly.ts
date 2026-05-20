import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0242 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-14 参数和字典 fallback 只读', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-14a: fallback rows hide direct edit and do not request missing details', async ({
    multiTenantMode,
    page,
  }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0242(page);
  });
});
