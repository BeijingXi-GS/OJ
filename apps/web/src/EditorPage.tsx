import Editor, { loader } from '@monaco-editor/react'
import * as monaco from 'monaco-editor/editor/editor.api.js'
import 'monaco-editor/languages/definitions/cpp/register.js'
import EditorWorker from 'monaco-editor/editor/editor.worker.js?worker'

self.MonacoEnvironment = { getWorker: () => new EditorWorker() }
loader.config({ monaco })

export default function EditorPage() {
  return (
    <>
      <p className="eyebrow">EDITOR PREVIEW</p>
      <h1>准备你的第一份代码。</h1>
      <p className="intro">
        仅用于本地编辑器检查。离开页面后内容不会保存；提交与判题将在后续阶段接入。
      </p>
      <div className="editor">
        <Editor
          height="440px"
          language="cpp"
          defaultValue={
            '#include <iostream>\n\nint main() {\n    std::cout << "Hello, OJ!" << std::endl;\n    return 0;\n}\n'
          }
          options={{
            minimap: { enabled: false },
            fontSize: 15,
            padding: { top: 20 },
            automaticLayout: true,
          }}
        />
      </div>
    </>
  )
}
