# 方案 B 规格：自绘 2.5D 校园地图 + GPS 仿射映射（前端）

> 适用项目：川农 120 周年「寻牛」微信小程序
> 本文档是前端开发的实现规格。后端配合事项见《后端接口开发注意清单》。

---

## 0. 一句话原理

地图不用任何地图服务商，就是**一张等比例 2.5D 渲染大图**。
通过 4~6 个实测控制点求一个**仿射矩阵**，运行时把任意 GCJ-02 经纬度换算成图上像素坐标，反向亦可。

```
经纬度(GCJ-02) → 局部平面坐标(米) → [仿射矩阵] → 大图像素坐标(px)
```

---

## 1. 组件架构

```
<CampusMap>                          ← 对业务层的唯一黑盒组件
 ├─ 手势层    拖拽 / 双指缩放 / 边界回弹（V1: movable-view，V2: canvas 自绘）
 ├─ 底图层    2.5D 大图（按缩放级别分瓦片，CDN 加载）
 ├─ 标记层    牛 Pin（彩色/剪影）、涟漪动画、用户蓝点
 └─ 映射模块  geo.ts（本规格 §4~§6 的纯函数，无任何框架依赖）
```

对外接口（与 Taro / uni-app / 原生无关）：

```ts
interface CampusMapProps {
  cows: CowMarker[];
  userLocation?: { lat: number; lng: number };   // GCJ-02
  focusCowId?: string;                            // 轮播联动：居中某头牛
  onCowTap(id: string): void;
}

/** 位置分级：未激活牛后端只下发模糊区域，精确坐标拿不到（玩法+防作弊要求） */
type CowMarker =
  | { id: string; activated: true;  lat: number; lng: number }                          // 精确点 → 彩色 Pin
  | { id: string; activated: false; area: { name: string; lat: number; lng: number; radiusM: number } };
      // 模糊区域 → 剪影画在区域圆心 + 半透明范围圈；卡片上的"距离 211m"= 用户到区域圆心的距离
```

---

## 2. 美术资源硬性要求（必须写进设计交付合同）

| 项 | 要求 | 原因 |
|---|---|---|
| 底图比例 | **地面布局严格按航拍正射影像等比例绘制**，不允许局部夸张放大某建筑占地 | 比例失真 = GPS 落点漂移，仿射救不回非线性形变 |
| 立体表现 | 2.5D 立体感只允许体现在建筑"长高"上；**每栋建筑的底座脚印位置必须与航拍对齐** | 锚点全部锚在底座，不锚楼顶 |
| 源图尺寸 | 长边 8192px（按校区实际长宽比），PNG/WebP 无损交付 | 支撑 3 级缩放不糊 |
| 瓦片切图 | 由前端用脚本切，设计只交整图；切 256×256，3 个缩放级别（1x/2x/4x） | — |
| 朝向 | 不强制正北朝上，**允许旋转**（仿射矩阵自带旋转项），但交付后不得再改构图比例 | 改图 = 重新标定 |

> 沟通纪要原话：「比例上要和真实地图对应哦，不然 GPS 落不准」——以上即其工程化落地条款。

---

## 3. 坐标系约定（全项目唯一坐标系：GCJ-02）

| 数据 | 坐标系 | 备注 |
|---|---|---|
| `wx.getLocation({ type: 'gcj02' })` | GCJ-02 | 运行时用户位置 |
| 控制点实测 | GCJ-02 | 用腾讯坐标拾取器或小程序自身 API 采集 |
| 牛坐标（数据库） | GCJ-02 | 用 §7 的录入工具生成 |
| 航拍图 / 无人机原始数据 | **WGS-84 ⚠️** | 仅供美术参考构图，**坐标值不得直接入库**，如需使用必须先 WGS-84 → GCJ-02 转换 |

**红线：任何一处混入 WGS-84 原始值，全图整体漂 300~600 米。**

---

## 4. 控制点采集 SOP（交给现场勘定人员）

1. 在大图上挑 **6 个**图上能精确点到 1px 的地标：推荐 校门门柱、图书馆建筑角、操场圆心、标志雕塑 等，**尽量分布在校园四角 + 中心**，不要集中在一侧
2. 每个点采集两份数据：
   - `pixel`: 在源图（8192px 版本）上的像素坐标 (px, py) —— 设计稿软件里取
   - `geo`: 现场实测 GCJ-02 (lat, lng) —— 人站到地标正下方，用采集工具读数，**同一点读 3 次取平均**（手机 GPS 单次误差 ±5~10m）
3. 6 个点中：**4 个用于求解，2 个留作验证**（验证点不参与求解）
4. 产出物：`calibration.json`

```json
{
  "imageSize": { "w": 8192, "h": 6144 },
  "origin": { "lat0": 30.7050, "lng0": 103.8650 },
  "points": [
    { "name": "东校门", "lat": 30.70512, "lng": 103.86618, "px": 7120, "py": 3050, "use": "solve" },
    { "name": "图书馆西北角", "lat": 30.70633, "lng": 103.86402, "px": 4480, "py": 1820, "use": "solve" },
    { "name": "操场圆心", "lat": 30.70411, "lng": 103.86330, "px": 3900, "py": 4310, "use": "solve" },
    { "name": "校史馆门口", "lat": 30.70588, "lng": 103.86205, "px": 2210, "py": 2400, "use": "solve" },
    { "name": "西门门柱", "lat": 30.70520, "lng": 103.86010, "px": 600,  "py": 3100, "use": "verify" },
    { "name": "钟楼", "lat": 30.70700, "lng": 103.86500, "px": 5400, "py": 800,  "use": "verify" }
  ]
}
```

---

## 5. 仿射矩阵求解（构建期执行一次，产物随包发布）

### 5.1 数学模型

第一步，经纬度转局部平面米制坐标（校园尺度 <2km，等距圆柱近似足够）：

```
x = (lng - lng0) × cos(lat0 · π/180) × 111320     // 米
y = (lat - lat0) × 110540                          // 米
```

第二步，平面坐标 → 像素，6 参数仿射（含 平移/旋转/缩放/轻微剪切）：

```
px = a·x + b·y + tx
py = c·x + d·y + ty
```

≥3 个控制点即可解；4 点用最小二乘，可吸收单点 GPS 误差。

### 5.2 参考实现（TypeScript，零依赖，直接进 `geo.ts`）

```ts
export interface Affine { a: number; b: number; c: number; d: number; tx: number; ty: number }

const R_LNG = 111320, R_LAT = 110540;

export function geoToPlane(lat: number, lng: number, lat0: number, lng0: number) {
  return {
    x: (lng - lng0) * Math.cos(lat0 * Math.PI / 180) * R_LNG,
    y: (lat - lat0) * R_LAT,
  };
}

/** 最小二乘求仿射：plane(x,y)[] → pixel(px,py)[]，points.length >= 3 */
export function solveAffine(
  pts: { x: number; y: number; px: number; py: number }[]
): Affine {
  // 两个独立的 3 参数最小二乘：A·[a,b,tx]ᵀ = px ；A·[c,d,ty]ᵀ = py
  // 正规方程 (AᵀA)θ = Aᵀb，AᵀA 为 3×3，高斯消元即可
  const AtA = [[0,0,0],[0,0,0],[0,0,0]];
  const AtPx = [0,0,0], AtPy = [0,0,0];
  for (const p of pts) {
    const row = [p.x, p.y, 1];
    for (let i = 0; i < 3; i++) {
      for (let j = 0; j < 3; j++) AtA[i][j] += row[i] * row[j];
      AtPx[i] += row[i] * p.px;
      AtPy[i] += row[i] * p.py;
    }
  }
  const [a, b, tx] = solve3x3(AtA, AtPx);
  const [c, d, ty] = solve3x3(AtA, AtPy);
  return { a, b, c, d, tx, ty };
}

function solve3x3(M: number[][], v: number[]): number[] {
  const m = M.map((r, i) => [...r, v[i]]);
  for (let col = 0; col < 3; col++) {
    let piv = col;
    for (let r = col + 1; r < 3; r++) if (Math.abs(m[r][col]) > Math.abs(m[piv][col])) piv = r;
    [m[col], m[piv]] = [m[piv], m[col]];
    for (let r = 0; r < 3; r++) {
      if (r === col) continue;
      const f = m[r][col] / m[col][col];
      for (let k = col; k < 4; k++) m[r][k] -= f * m[col][k];
    }
  }
  return [m[0][3] / m[0][0], m[1][3] / m[1][1], m[2][3] / m[2][2]];
}
```

### 5.3 运行时正/反变换

```ts
/** GCJ-02 → 图上像素（画牛 Pin、用户蓝点） */
export function geoToPixel(lat: number, lng: number, cal: Calibration): { px: number; py: number } {
  const { x, y } = geoToPlane(lat, lng, cal.origin.lat0, cal.origin.lng0);
  const t = cal.affine;
  return { px: t.a * x + t.b * y + t.tx, py: t.c * x + t.d * y + t.ty };
}

/** 图上像素 → GCJ-02（仅管理端录牛坐标用） */
export function pixelToGeo(px: number, py: number, cal: Calibration): { lat: number; lng: number } {
  const t = cal.affine;
  const det = t.a * t.d - t.b * t.c;
  const dx = px - t.tx, dy = py - t.ty;
  const x = ( t.d * dx - t.b * dy) / det;
  const y = (-t.c * dx + t.a * dy) / det;
  return {
    lat: cal.origin.lat0 + y / R_LAT,
    lng: cal.origin.lng0 + x / (Math.cos(cal.origin.lat0 * Math.PI / 180) * R_LNG),
  };
}
```

### 5.4 验收标准（标定完成的定义）

| 项 | 标准 |
|---|---|
| 2 个验证点换算误差 | 像素误差换算回米 **≤ 10m**（误差米数 = 像素距离 ÷ 比例尺，比例尺 = √\|det\|） |
| 现场走测 | 持手机沿主干道走一圈，蓝点不出路面贴线漂移 |
| 失败处理 | 误差 >10m：先查坐标系混用（§3），再查美术图比例失真（§2），最后才加控制点重解 |

求解脚本作为 `scripts/calibrate.ts` 进仓库，输出 `calibration.json`（仅几百字节）。

**分发方式：标定数据由后端 `/config` 接口下发，小程序包内置一份作为兜底缓存。** 这样美术改图、重新标定时只需更新服务端配置，**不用发小程序版（审核 1~3 天）**。`calibration.json` 带版本号字段，前端以 `/config` 返回的为准。

---

## 6. 地图交互实现（两阶段）

### V1（首发版）：`movable-view`
- `movable-area` 固定视口 + `movable-view` 承载大图，`scale` 开双指缩放（0.5x~2x）
- 牛 Pin / 蓝点为 movable-view 内部的绝对定位节点，位置 = `geoToPixel` 结果（像素按当前底图显示尺寸/源图尺寸比例缩放）
- 涟漪、剪影→彩色切换用 CSS animation
- 轮播联动：点底部牛卡片 → 计算该牛像素坐标，设置 movable-view 偏移使其居中（带 `animation: true`）

### V2（仅当 V1 真机卡顿才做）：canvas 2d 自绘
- type=2d canvas + 自管 touch 手势，瓦片按视口懒加载
- 触发条件：标记 >30 个同屏、或低端机（红米/荣耀百元机）拖拽掉帧 <40fps

### 性能红线
1. 拖拽过程**禁止 setData 驱动位移**——movable-view 原生手势自己跑，JS 只在 `bindchange` 节流读位置
2. 同屏牛 Pin 按策划要求轮播露出 5~10 个，**不要 120 个全挂 DOM**
3. 瓦片与皮肤序列帧全部 CDN 远程加载，主包不放任何大图
4. 序列帧动画用雪碧图 + `steps()`，不要 100 张 img 换 src

---

## 7. 牛坐标录入工具（管理端，1 天工作量）

- 一个隐藏页面/独立 H5：展示同一张大图，管理员**点击图上位置 → `pixelToGeo` 反解经纬度 → 调后端 `/admin/cows` 录入**
- 录入时一并填写：**区域名**（如"图书馆附近"，用于未激活牛的模糊展示）、上线规则、是否铁牛加成；模糊圆心由**后端**生成并固化（前端工具不管）
- 比现场抄录 120 个坐标可靠一个量级；现场勘定人员只需确认"牛放在图上哪里"
- 录完后可一键导出全部牛坐标 KML/JSON 给主办方核对

---

## 8. 风险与降级预案

| 风险 | 表现 | 预案 |
|---|---|---|
| 美术图比例失真且无法返工 | 验证点误差压不进 10m | 降级路线 A：原生 `<map>` + groundoverlay 贴图，牺牲视觉保 GPS（业务层接口不变，只换 `<CampusMap>` 内部实现） |
| 单点 GPS 噪声 | 个别控制点误差大 | 已留 2 验证点交叉检查；重测该点 3 次取平均 |
| 用户 GPS 漂移 | 蓝点抖动 | 前端对连续定位做滑动平均；激活判距在服务端完成（见后端清单） |
