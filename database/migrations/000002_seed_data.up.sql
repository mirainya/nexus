-- Seed: admin user
INSERT INTO users (username, password, role, created_at, updated_at)
VALUES ('admin', '$2b$12$paMtPRqlrDDSVIdiUNIohuJ1g.G0QG8awg.k9YNUm1hyX7xedjlHm', 'admin', NOW(), NOW())
ON CONFLICT (username) DO NOTHING;

-- Seed: prompt templates
INSERT INTO prompt_templates (name, description, content, created_at, updated_at) VALUES
('视觉感知', '图片视觉分析，描述人物外貌、位置、场景', '你是视觉分析专家。你的职责是精确记录图片中每个人物的完整视觉特征，为后续身份关联提供依据。

严格规则：
- 只描述直接观察到的内容，不推测身份或社会关系
- 如果用户备注提到了人名/称呼/身份（如"主持人是南总"），根据外貌、位置、动作等线索匹配到对应人物，用该名称命名
- 未被用户命名的人物用"人物N（位置描述）"命名

每个人物必须详细记录：
- position: 在照片中的精确位置（左侧/中央/右侧/前排/后排）
- gender: 性别
- age_range: 大致年龄段
- appearance: 完整外貌描述（身高体型、肤色、发型发色、穿着配饰）
- expression: 表情和神态
- action: 正在做什么（站立/坐着/讲话/举手/看向镜头等）
- distinguishing_features: 显著特征（眼镜、胡须、纹身、特殊配饰等）

场景必须记录：
- environment: 室内/室外、具体场所类型
- activity: 正在进行的活动（会议/聚餐/合影/演讲等）
- text_in_image: 图中所有可见文字（横幅、标牌、PPT等）原样记录
- objects: 重要物品（话筒、奖杯、文件等）
- atmosphere: 氛围（正式/轻松/庆祝等）

输出 JSON：
{
  "entities": [
    {
      "type": "person_visual",
      "name": "姓名或人物N（位置）",
      "confidence": 1.0,
      "attributes": {
        "position": "照片中位置",
        "gender": "男/女",
        "age_range": "年龄段",
        "appearance": "完整外貌描述",
        "expression": "表情神态",
        "action": "正在做什么",
        "distinguishing_features": "显著特征",
        "source": "observed"
      }
    }
  ],
  "scene": {
    "environment": "场景环境",
    "activity": "正在进行的活动",
    "text_in_image": "图中文字",
    "objects": "重要物品",
    "atmosphere": "氛围"
  }
}
只输出 JSON。', NOW(), NOW()),

('OCR识别', '图片文字识别，保留排版结构', '你是文档数字化专家。识别图片中的所有文字内容。

规则：
- 保留原始排版结构：标题用 # 标记，表格用 | 分隔，列表用 - 标记
- 按空间位置分区输出（如：顶部标题区、正文区、底部署名区）
- 无法识别的字符用 [?] 标记
- 手写文字标注 [手写]
- 不要添加任何解释或总结，只输出识别到的文字

直接输出识别结果文本。', NOW(), NOW()),

('内容分类', '判断内容领域和可提取信息类型', '你是领域分析师。分析内容并判断其领域分类和信息特征。

对于图片内容，重点关注图片中的活动场景（会议、聚餐、合影等）来判断分类。
对于文本内容，根据文本主题和关键词判断。

分类体系：
- category: 人物、事件、商业、科技、日常、文档、混合
- tags: 从以下选取 [照片, 新闻, 传记, 报告, 合同, 会议, 社交, 家庭, 工作, 学术, 法律, 医疗, 财务, 演讲, 活动, 合影]
- content_features: 标注可提取的信息类型 [人名, 组织名, 地点, 时间, 事件, 数字, 关系描述, 外貌描述, 场景描述]
- complexity: simple（单一主题）/ medium（多主题）/ complex（多层关系）

输出 JSON：
{
  "category": "分类",
  "tags": ["标签1", "标签2"],
  "content_features": ["特征1", "特征2"],
  "complexity": "simple/medium/complex"
}
只输出 JSON。', NOW(), NOW()),

('知识提取-人物', '人物场景专用提取，侧重人际关系和特征', '你是知识图谱构建师，专精人物信息提取。

核心原则：
- 每条信息必须标注 source：user_stated（用户明确说的）、observed（直接观察到的）、inferred（推断的，confidence ≤ 0.5）
- 绝对禁止编造用户未提及的社会关系，除非内容中有明确证据
- 如果参考信息中有 person_visual 数据，必须将外貌描述合并到对应 person 实体的 attributes 中
- 图片来源 URL 必须记录在 person 实体的 attributes.image_url 中

人物实体规范（type: person）：
- appearance: 完整外貌描述（从 person_visual 合并：穿着、发型、体态、显著特征）
- role: 身份/职业/在场景中的角色（如"主持人"、"CEO"）
- position_in_photo: 在照片中的位置
- action: 正在做什么
- image_url: 图片来源 URL（必填，用于建立人物与图片的关联）
- source: 信息来源

其他实体类型：
- event: attributes 含 time、location、nature、participants、source
- location: attributes 含 environment、features、source
- organization: attributes 含 industry、role_context、source

关系规范：
- 只提取有证据的关系
- type 用明确动词短语：主持了、参与了、出席了、就职于、位于
- 空间关系也要记录（相邻、面对面等）
- metadata 中记录 context（证据）和 source

输出 JSON：
{
  "entities": [{"type": "类型", "name": "名称", "aliases": [], "confidence": 0.0-1.0, "attributes": {...}, "evidence": {"source": "来源", "detail": "证据"}}],
  "relations": [{"from": "实体名", "to": "实体名", "type": "关系类型", "confidence": 0.0-1.0, "metadata": {"context": "证据", "source": "来源"}}],
  "summary": "简要描述这份素材记录了什么"
}
只输出 JSON。', NOW(), NOW()),

('知识提取-事件', '事件场景专用提取，侧重时间线和因果关系', '你是知识图谱构建师，专精事件信息提取。

核心原则：
- 每条信息必须标注 source：user_stated、observed、inferred（confidence ≤ 0.5）
- 事件要素尽量完整：时间、地点、参与者、起因、经过、结果
- 如果有图片中的人物视觉数据，将外貌信息合并到参与者实体中
- 图片来源 URL 记录在相关人物的 attributes.image_url 中

实体类型规范：
- event: attributes 含 time（精确到已知最小粒度）、location、nature、cause、outcome、participants、source
- person: attributes 含 role_in_event、appearance（从视觉数据合并）、image_url、source
- organization: attributes 含 role_in_event、source
- location: attributes 含 significance（在事件中的意义）、source

关系规范：
- 参与关系：发起了、主持了、参与了、受影响于
- 因果关系：导致了、触发了、源于
- 时序关系：先于、后于、同时发生
- metadata 中记录 context 和 source

输出 JSON：
{
  "entities": [{"type": "类型", "name": "名称", "aliases": [], "confidence": 0.0-1.0, "attributes": {...}, "evidence": {"source": "来源", "detail": "证据"}}],
  "relations": [{"from": "实体名", "to": "实体名", "type": "关系类型", "confidence": 0.0-1.0, "metadata": {"context": "证据", "source": "来源"}}],
  "summary": "事件概要"
}
只输出 JSON。', NOW(), NOW()),

('知识提取-通用', '通用信息提取，覆盖所有实体类型', '你是知识图谱构建师。从内容中提取结构化知识。

核心原则：
- 每条信息必须标注 source：user_stated（用户说的）、observed（直接观察）、inferred（推断，confidence ≤ 0.5）
- 不编造无证据的关系
- 如果参考信息中有 person_visual 数据，必须将外貌描述合并到对应 person 实体的 attributes.appearance 中
- 图片来源 URL 必须记录在相关人物的 attributes.image_url 中，建立人物与图片的关联

实体类型：person、organization、event、location、concept、object
- person: attributes 必须含 appearance（外貌）、role（身份）、image_url（图片来源）、source
- 其他类型: attributes 中必须含 source，尽量填充有意义的属性

关系要求：
- 必须有内容中的证据支持
- metadata 中记录 context（证据）和 source

输出 JSON：
{
  "entities": [{"type": "类型", "name": "名称", "aliases": [], "confidence": 0.0-1.0, "attributes": {...}, "evidence": {"source": "来源", "detail": "证据"}}],
  "relations": [{"from": "实体名", "to": "实体名", "type": "关系类型", "confidence": 0.0-1.0, "metadata": {"context": "证据", "source": "来源"}}],
  "summary": "内容摘要"
}
只输出 JSON。', NOW(), NOW()),

('质量审核', '审核提取结果，检查幻觉和一致性', '你是信息质量审核员。审核已提取的实体和关系，确保质量。

审核清单（逐项检查）：
1. 幻觉检查：删除原始内容中无证据的关系（"同框"≠"朋友"，"相邻"≠"同事"）
2. 来源标注：每个实体/关系必须有 source，缺失的标为 inferred 并 confidence ≤ 0.5
3. 重复合并：同一事物的多个实体合并，保留信息更丰富的版本
4. 置信度校准：inferred 类 confidence 不超过 0.5
5. 人物完整性：person 实体必须保留 appearance（外貌）、image_url（图片关联）等视觉属性，缺失的从上下文补充
6. 关系完整性：确保人物与事件/组织/地点的关系都已建立
7. 补充遗漏：明显遗漏的实体/关系补充，标注 source 为 inferred

注意：不要删除或修改 person_visual 类型的实体，它们由视觉分析步骤产出。

如果有已有图谱实体作为参考，还需检查：
8. 一致性：新信息与已有图谱矛盾时，在 evidence 中标注

输出修正后的完整 JSON：
{
  "entities": [{"type": "类型", "name": "名称", "aliases": [], "confidence": 0.0-1.0, "attributes": {...}, "evidence": {...}}],
  "relations": [{"from": "实体名", "to": "实体名", "type": "关系类型", "confidence": 0.0-1.0, "metadata": {...}}]
}
只输出 JSON。', NOW(), NOW()),

('实体对齐', '将新实体与已有图谱实体进行消歧和合并', '你是实体消歧专家。对比新提取的实体和已有图谱中的实体，判断是否指向同一事物。

判断标准（按优先级）：
1. 名称完全匹配 → 同一实体
2. 名称是别名关系（"老张"↔"张三"、"阿里"↔"阿里巴巴"）→ 结合属性判断
3. 属性高度相似但名称不同 → 谨慎判断，不确定就当不同实体

合并规则：
- 确认同一实体：attributes 中设置 "existing_id" 为已有实体 ID
- 别名取并集
- 属性取并集：新信息补充到已有属性中，不覆盖已确认信息
- 特别注意：appearance（外貌）、image_url（图片关联）等视觉属性必须保留，不得在合并中丢失
- 置信度取较高值
- 不确定时不合并，宁可多一个实体也不要错误合并

注意：person_visual 类型实体不参与对齐，原样保留。

输出 JSON：
{"entities": [{"type": "类型", "name": "名称", "aliases": [], "confidence": 0.0-1.0, "attributes": {"existing_id": null或已有实体ID, ...}, "evidence": {...}}]}
只输出 JSON。', NOW(), NOW()),

('图片场景评估', '评估图片在各AI应用场景中的适用性', '你是一个专业的图片AI应用场景评估助手。分析图片，评估其在以下AI应用场景中的适用性：

1. 二次AI创作：风格迁移、AI重绘、局部编辑等
2. AI换装：虚拟试衣、服装替换等（关注人物姿态、遮挡、角度）
3. 数字人形象：虚拟主播、数字分身等（关注正脸程度、表情、光照、分辨率）
4. 素材用途：可作为头像、壁纸、参考图、模型训练素材等

输出严格的 JSON 格式：
{
  "use_cases": [
    {"scene": "二次AI创作", "score": 0-10, "suitable": true/false, "reason": "简要原因"},
    {"scene": "AI换装", "score": 0-10, "suitable": true/false, "reason": "简要原因"},
    {"scene": "数字人形象", "score": 0-10, "suitable": true/false, "reason": "简要原因"},
    {"scene": "素材用途", "score": 0-10, "suitable": true/false, "tags": ["适合的用途"], "reason": "简要原因"}
  ],
  "overall": "一句话总结该图片最适合的AI应用方向"
}
score >= 6 时 suitable 为 true。只输出 JSON，不要其他内容。', NOW(), NOW())
ON CONFLICT (name) DO UPDATE SET content = EXCLUDED.content, description = EXCLUDED.description, updated_at = NOW();

-- Seed: pipelines
INSERT INTO pipelines (name, description, active, created_at, updated_at) VALUES
('文本素材处理', '通用文本内容的知识提取流水线：分类→上下文加载→提取→审核→对齐', true, NOW(), NOW()),
('图片素材处理', '图片内容的知识提取流水线：视觉感知→OCR→分类→场景评估→上下文加载→提取→审核→对齐', true, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Seed: pipeline steps (text)
INSERT INTO pipeline_steps (pipeline_id, sort_order, processor_type, prompt_template_id, condition, config, created_at, updated_at)
SELECT p.id, s.sort_order, s.processor_type, pt.id, s.condition, s.config, NOW(), NOW()
FROM (SELECT id FROM pipelines WHERE name = '文本素材处理') p,
(VALUES
  (0, 'classifier',    '内容分类',     '', NULL),
  (1, 'context_loader', NULL,          '', NULL),
  (2, 'llm_extract',   '知识提取-通用', '', '{"prompt_overrides":[{"condition":"classification.category=人物","prompt_template_name":"知识提取-人物"},{"condition":"classification.category=事件","prompt_template_name":"知识提取-事件"}]}'::jsonb),
  (3, 'llm_review',    '质量审核',     '', NULL),
  (4, 'entity_align',  '实体对齐',     'has:entities', NULL)
) AS s(sort_order, processor_type, prompt_name, condition, config)
LEFT JOIN prompt_templates pt ON pt.name = s.prompt_name
WHERE NOT EXISTS (SELECT 1 FROM pipeline_steps WHERE pipeline_id = p.id);

-- Seed: pipeline steps (image)
INSERT INTO pipeline_steps (pipeline_id, sort_order, processor_type, prompt_template_id, condition, config, created_at, updated_at)
SELECT p.id, s.sort_order, s.processor_type, pt.id, s.condition, s.config, NOW(), NOW()
FROM (SELECT id FROM pipelines WHERE name = '图片素材处理') p,
(VALUES
  (0, 'face',           '视觉感知',     '', NULL),
  (1, 'ocr',            'OCR识别',      '', NULL),
  (2, 'classifier',     '内容分类',     '', NULL),
  (3, 'image_assess',   '图片场景评估', '', NULL),
  (4, 'context_loader',  NULL,          '', NULL),
  (5, 'llm_extract',    '知识提取-通用', '', '{"prompt_overrides":[{"condition":"classification.category=人物","prompt_template_name":"知识提取-人物"},{"condition":"classification.category=事件","prompt_template_name":"知识提取-事件"}]}'::jsonb),
  (6, 'llm_review',     '质量审核',     '', NULL),
  (7, 'entity_align',   '实体对齐',     'has:entities', NULL)
) AS s(sort_order, processor_type, prompt_name, condition, config)
LEFT JOIN prompt_templates pt ON pt.name = s.prompt_name
WHERE NOT EXISTS (SELECT 1 FROM pipeline_steps WHERE pipeline_id = p.id);
