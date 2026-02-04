# ============================================
# Stage 1: 安装依赖 (利用缓存)
# ============================================
FROM node:20-alpine AS deps

RUN apk add --no-cache python3 make g++

WORKDIR /app

# 只复制依赖相关文件
COPY package*.json .npmrc ./
RUN npm install --ignore-scripts

# ============================================
# Stage 2: 构建
# ============================================
FROM node:20-alpine AS builder

WORKDIR /app

# 从 deps 阶段复制 node_modules
COPY --from=deps /app/node_modules ./node_modules

# 复制源代码（.dockerignore 排除不需要的文件）
COPY . .

# 应用补丁（如果有）
RUN npx patch-package || true

# 构建 (增加 Node 内存限制到 8GB)
ENV NODE_OPTIONS="--max-old-space-size=8192"
RUN npx cross-env CHATBOX_BUILD_PLATFORM=web NODE_ENV=production \
    TS_NODE_TRANSPILE_ONLY=true API_BASE_URL= \
    webpack --config ./.erb/configs/webpack.config.renderer.prod.ts

# ============================================
# Stage 3: 生产镜像
# ============================================
FROM nginx:alpine

COPY --from=builder /app/release/app/dist/renderer /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
