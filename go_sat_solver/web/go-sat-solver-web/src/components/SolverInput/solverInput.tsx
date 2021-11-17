import React from 'react';
import MonacoEditor from 'react-monaco-editor';

export const SolverInput: React.FunctionComponent<{
    input: string;
    onInputChange: (newInput: string) => void;
}> = ({ input, onInputChange, }) => {

    const onEditorDidMount = (editor: any) => {
        editor.focus();
    };

    const onEditorWillMount = (monaco: any) => {

        // Register a new language
        monaco.languages.register({ id: 'sat-haskell' });

        // Register a tokens provider for the language
        monaco.languages.setMonarchTokensProvider('sat-haskell', {
            tokenizer: {
                root: [
                    [/And/, "and"],
                    [/Or/, "or"],
                    [/Not/, "not"],
                    [/T/, "true"],
                    [/F/, "false"],
                    [/Var "[a-zA-Z0-9_-]+"/, "var"],
                    [/[()]/, 'bracket'],
                ]
            }
        });

        // Define a new theme that contains only rules that match this language
        monaco.editor.defineTheme('sat-haskell-theme', {
            base: 'vs-dark',
            inherit: false,
            rules: [
                { token: 'var', foreground: '3D9970' },
                { token: 'false', foreground: 'ff0000' },
                { token: 'true', foreground: '3D9970' },
                { token: 'not', foreground: 'ff0000' },
                { token: 'or', foreground: '0074D9' },
                { token: 'and', foreground: '0074D9' },
                { token: 'bracket', foreground: 'AAAAAA' },
                { foreground: 'FFFFFF' }
            ],
            colors: {
                "editor.background": '#282c34',
                "editor.foreground": '#EEEEEE',
                "editorCursor.foreground": '#EEEEEE',
            }
        });

        // Register a completion item provider for the new language
        monaco.languages.registerCompletionItemProvider('sat-haskell', {
            provideCompletionItems: () => {
                const suggestions: any = [];
                return { suggestions: suggestions };
            }
        });

    };

    return (
        <div style={{ textAlign: 'left', }}>
            <MonacoEditor
                width="800"
                height="600"
                language="sat-haskell"
                theme="sat-haskell-theme"
                value={input}
                options={{
                    automaticLayout: true,
                    cursorBlinking: "smooth",
                    wordWrap: "on",
                    wrappingIndent: "same",
                    scrollBeyondLastLine: false,
                    minimap: {
                        enabled: false,
                    },
                    fontSize: 20,
                }}
                onChange={onInputChange}
                editorDidMount={onEditorDidMount}
                editorWillMount={onEditorWillMount}
            />
        </div>
    );
};
