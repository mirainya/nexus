-- 为 pipeline_steps 添加 parallel_group 列，支持同组步骤并行执行
ALTER TABLE pipeline_steps ADD COLUMN IF NOT EXISTS parallel_group INT NOT NULL DEFAULT 0;

-- 图片流水线：face/ocr/image_assess 设为同一并行组
UPDATE pipeline_steps SET parallel_group = 1
WHERE pipeline_id = (SELECT id FROM pipelines WHERE name = '图片素材处理')
  AND processor_type IN ('face', 'ocr', 'image_assess');
