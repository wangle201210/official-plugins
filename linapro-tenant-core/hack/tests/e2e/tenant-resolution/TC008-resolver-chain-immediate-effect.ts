import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0205 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-4 解析链固定策略', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-4a: removed resolver policy API leaves code-owned policy unchanged', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0205();
  });
});
