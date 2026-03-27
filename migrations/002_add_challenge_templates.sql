-- AWD Arena Platform - Challenge Templates
-- PostgreSQL 17

-- Challenge Templates
CREATE TABLE challenge_templates (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(200)  NOT NULL UNIQUE,
    category        VARCHAR(50)   NOT NULL,
    description     TEXT,
    
    -- Docker镜像配置
    image_config    JSONB         NOT NULL DEFAULT '{}',
    
    -- 服务端口配置
    service_ports   JSONB         NOT NULL DEFAULT '[]',
    
    -- 漏洞配置
    vuln_config     JSONB         NOT NULL DEFAULT '{}',
    
    -- Flag配置
    flag_config     JSONB         NOT NULL DEFAULT '{}',
    
    -- 难度和分数
    difficulty      VARCHAR(20)   NOT NULL DEFAULT 'medium',
    base_score      INT           NOT NULL DEFAULT 100,
    
    -- 资源限制
    cpu_limit       DECIMAL(3,2)  NOT NULL DEFAULT 0.5,
    mem_limit       INT           NOT NULL DEFAULT 256,
    
    -- 提示信息
    hints           TEXT,
    
    -- 状态
    status          VARCHAR(20)   NOT NULL DEFAULT 'draft',
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_challenge_templates_category ON challenge_templates(category);
CREATE INDEX idx_challenge_templates_difficulty ON challenge_templates(difficulty);
CREATE INDEX idx_challenge_templates_status ON challenge_templates(status);

-- 注释
COMMENT ON TABLE challenge_templates IS '题库模板表';
COMMENT ON COLUMN challenge_templates.name IS '模板名称';
COMMENT ON COLUMN challenge_templates.category IS '题目类别：web, pwn, crypto, reverse, misc';
COMMENT ON COLUMN challenge_templates.image_config IS 'Docker镜像配置：包含镜像名称、标签、环境变量等';
COMMENT ON COLUMN challenge_templates.service_ports IS '服务端口配置数组';
COMMENT ON COLUMN challenge_templates.vuln_config IS '漏洞配置：类型、CVE、修复方案等';
COMMENT ON COLUMN challenge_templates.flag_config IS 'Flag配置：类型、位置、生成规则等';
COMMENT ON COLUMN challenge_templates.difficulty IS '难度：easy, medium, hard';
COMMENT ON COLUMN challenge_templates.status IS '状态：draft, published, archived';

