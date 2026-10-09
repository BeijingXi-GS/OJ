import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { getDebug } from '@/lib/api'
import { useMode } from '@/lib/mode'
import { formatMemory } from '@/lib/utils'

export default function DebugPage() {
  const mode = useMode()
  const debug = useQuery({
    queryKey: ['debug', mode],
    queryFn: ({ signal }) => getDebug(mode, signal),
    enabled: mode === 'engineer',
  })
  if (mode !== 'engineer') return null
  return (
    <div className="debug-page">
      <Link className="back-link" to="/">
        ← 返回题库
      </Link>
      <div className="page-heading">
        <p className="eyebrow">ENGINEERING</p>
        <h1>工程调试</h1>
        <p>
          查看服务配置与判题元数据。运行的退出码、OOM
          事件和追溯信息显示在题目测试结果中。
        </p>
      </div>
      <section
        className="panel debug-service"
        aria-labelledby="debug-service-title"
      >
        <div className="panel-heading">
          <h2 id="debug-service-title">服务检查</h2>
          <Button
            variant="outline"
            onClick={() => void debug.refetch()}
            disabled={debug.isFetching}
          >
            {debug.isFetching ? '检查中…' : '刷新检查'}
          </Button>
        </div>
        {debug.isPending ? (
          <p className="load-card" role="status">
            正在读取配置与元数据…
          </p>
        ) : debug.isError ? (
          <div className="request-error" role="alert">
            <h3>检查未完成</h3>
            <p>{debug.error.message}</p>
          </div>
        ) : (
          <div className="debug-service-body">
            <p className="debug-explanation">{debug.data.message}</p>
            <dl className="debug-values">
              <div>
                <dt>运行范围</dt>
                <dd>
                  {debug.data.local_demo ? '本机演示 · 手动切换模式' : '未知'}
                </dd>
              </div>
              <div>
                <dt>编译服务配置</dt>
                <dd>{debug.data.compiler_configured ? '已配置' : '未配置'}</dd>
              </div>
              <div>
                <dt>判题服务配置</dt>
                <dd>{debug.data.judge_configured ? '已配置' : '未配置'}</dd>
              </div>
              <div>
                <dt>登记题目</dt>
                <dd>{debug.data.problem_count} 道</dd>
              </div>
              <div>
                <dt>判题元数据读取</dt>
                <dd>
                  {debug.data.judge_metadata.available
                    ? '本次读取成功'
                    : '本次不可用'}
                </dd>
              </div>
              {debug.data.judge_metadata.available ? (
                <>
                  <div>
                    <dt>检查题目</dt>
                    <dd>{debug.data.judge_metadata.problem_id}</dd>
                  </div>
                  <div>
                    <dt>文件用例</dt>
                    <dd>{debug.data.judge_metadata.case_count} 个</dd>
                  </div>
                  <div>
                    <dt>每例总耗时限额</dt>
                    <dd>{debug.data.judge_metadata.wall_time_ms} ms</dd>
                  </div>
                  <div>
                    <dt>每例内存限额</dt>
                    <dd>
                      {debug.data.judge_metadata.memory_limit_bytes !==
                      undefined
                        ? formatMemory(
                            debug.data.judge_metadata.memory_limit_bytes,
                          )
                        : '未提供'}
                    </dd>
                  </div>
                  <div>
                    <dt>比较方式</dt>
                    <dd>{debug.data.judge_metadata.checker}</dd>
                  </div>
                </>
              ) : (
                debug.data.judge_metadata.error_code && (
                  <div>
                    <dt>元数据错误代码</dt>
                    <dd>{debug.data.judge_metadata.error_code}</dd>
                  </div>
                )
              )}
            </dl>
            <div className="debug-next-step">
              <h3>验证一次实际运行</h3>
              <p>
                打开一道题目，在工程师模式中运行测试。展开用例可以查看退出码、OOM
                事件和通信方向；展开工程运行证据可以查看本次源码、数据、产物和镜像标识。
              </p>
              <Button asChild variant="outline">
                <Link to="/problems/a-plus-b">打开 A + B</Link>
              </Button>
            </div>
          </div>
        )}
      </section>
    </div>
  )
}
