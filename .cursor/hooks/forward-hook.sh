#!/bin/bash
# Hook 转发 + 需求状态管理 + 上下文注入
PIPELINE_SERVER="http://127.0.0.1:19091"
REQ_DIR="requirements/in-progress"
input=$(cat)

event_name=$(echo "$input" | grep -o '"hook_event_name":"[^"]*"' | head -1 | cut -d'"' -f4)

# sessionStart: 注册模板 + 注入活跃需求上下文
if [ "$event_name" = "sessionStart" ]; then
  pipelines_dir="$(pwd)/.cursor/pipelines"
  if [ -d "$pipelines_dir" ]; then
    project_name=$(basename "$(pwd)")
    curl -s --max-time 2 -X POST "$PIPELINE_SERVER/api/v1/templates/register" \
      -H "Content-Type: application/json" \
      -d "{\"project\":\"$project_name\",\"dir\":\"$pipelines_dir\"}" >/dev/null 2>&1 || true
  fi

  ctx=""
  if [ -d "$REQ_DIR" ]; then
    for status_file in "$REQ_DIR"/*/.status; do
      [ -f "$status_file" ] || continue
      req_id=$(basename "$(dirname "$status_file")")
      stage=$(cat "$status_file")
      readme="$(dirname "$status_file")/README.md"
      title=""
      if [ -f "$readme" ]; then
        title=$(head -1 "$readme" | sed 's/^# //')
      fi
      ctx="${ctx}[需求 ${req_id}] 阶段: ${stage}"
      [ -n "$title" ] && ctx="${ctx} | ${title}"
      ctx="${ctx}\n"
    done
  fi

  response=$(echo "$input" | curl -s --max-time 3 -X POST -H "Content-Type: application/json" -d @- "$PIPELINE_SERVER/api/v1/pipeline/hook" 2>/dev/null)
  if [ $? -ne 0 ] || [ -z "$response" ]; then
    if [ -n "$ctx" ]; then
      echo "{\"additional_context\":\"$(echo -e "$ctx" | sed 's/"/\\"/g')\"}"
    else
      echo '{}'
    fi
    exit 0
  fi

  if [ -n "$ctx" ]; then
    existing_ctx=$(echo "$response" | grep -o '"additional_context":"[^"]*"' | head -1 | cut -d'"' -f4)
    merged="${existing_ctx}\\n${ctx}"
    response=$(echo "$response" | sed "s|\"additional_context\":\"[^\"]*\"|\"additional_context\":\"${merged}\"|")
    if ! echo "$response" | grep -q "additional_context"; then
      response=$(echo "$response" | sed 's/}$/,"additional_context":"'"$(echo -e "$ctx" | sed 's/"/\\"/g')"'"}/')
    fi
  fi
  echo "$response"
  exit 0
fi

# stop: 注入知识沉淀提示 + 阶段上下文
if [ "$event_name" = "stop" ]; then
  ctx=""
  if [ -d "$REQ_DIR" ]; then
    for status_file in "$REQ_DIR"/*/.status; do
      [ -f "$status_file" ] || continue
      req_id=$(basename "$(dirname "$status_file")")
      stage=$(cat "$status_file")
      ctx="${ctx}[活跃需求 ${req_id}] 当前阶段: ${stage}。"

      case "$stage" in
        brainstorm)
          ctx="${ctx} 下一步: design-gate（展示 brainstorm 结果等待确认）。"
          ;;
        design-gate)
          ctx="${ctx} 下一步: plan（调用 ce:plan 进行技术规划）。"
          ;;
        plan)
          ctx="${ctx} 下一步: plan-gate（展示技术方案等待确认）。"
          ;;
        plan-gate)
          ctx="${ctx} 下一步: implement（按方案编码实现）。"
          ;;
        implement)
          ctx="${ctx} 下一步: code-review（调用 ce:review 代码审查）。"
          ;;
        code-review)
          ctx="${ctx} 下一步: code-gate（展示审查结果等待确认）。"
          ;;
        code-gate)
          ctx="${ctx} 下一步: unit-test（编写单元测试）。"
          ;;
        unit-test)
          ctx="${ctx} 下一步: test-plan-gate（展示测试覆盖等待确认）。"
          ;;
        test-plan-gate)
          ctx="${ctx} 下一步: e2e-test（执行端到端测试）。"
          ;;
        e2e-test)
          ctx="${ctx} 下一步: test-gate（展示 E2E 结果等待确认）。"
          ;;
        test-gate)
          ctx="${ctx} 下一步: archive（归档 + 知识沉淀）。"
          ;;
        reproduce-test)
          ctx="${ctx} 下一步: test-gate（确认 Bug 已复现）。"
          ;;
        fix)
          ctx="${ctx} 下一步: unit-test-pass（验证修复，复现测试应 PASS）。"
          ;;
        unit-test-pass)
          ctx="${ctx} 下一步: archive（归档 + 知识沉淀）。"
          ;;
      esac
      ctx="${ctx} 提示: 如有值得沉淀的知识，使用 /optimize-flow 或在 archive 阶段自动沉淀。\n"
    done
  fi

  response=$(echo "$input" | curl -s --max-time 3 -X POST -H "Content-Type: application/json" -d @- "$PIPELINE_SERVER/api/v1/pipeline/hook" 2>/dev/null)
  if [ $? -ne 0 ] || [ -z "$response" ]; then
    if [ -n "$ctx" ]; then
      echo "{\"additional_context\":\"$(echo -e "$ctx" | sed 's/"/\\"/g')\"}"
    else
      echo '{}'
    fi
    exit 0
  fi

  if [ -n "$ctx" ]; then
    echo "{\"additional_context\":\"$(echo -e "$ctx" | sed 's/"/\\"/g')\"}"
  else
    echo "$response"
  fi
  exit 0
fi

# 其他事件: 直接转发给 pipeline-server
response=$(echo "$input" | curl -s --max-time 3 -X POST -H "Content-Type: application/json" -d @- "$PIPELINE_SERVER/api/v1/pipeline/hook" 2>/dev/null)
[ $? -ne 0 ] || [ -z "$response" ] && echo '{}' && exit 0
echo "$response"
