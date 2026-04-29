DELETE FROM pipeline_steps WHERE pipeline_id IN (SELECT id FROM pipelines WHERE name IN ('文本素材处理', '图片素材处理'));
DELETE FROM pipelines WHERE name IN ('文本素材处理', '图片素材处理');
DELETE FROM prompt_templates WHERE name IN ('视觉感知', 'OCR识别', '内容分类', '知识提取-人物', '知识提取-事件', '知识提取-通用', '质量审核', '实体对齐', '图片场景评估');
DELETE FROM users WHERE username = 'admin';
