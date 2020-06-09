import React from 'react';
import {INITIAL_SETTINGS, SolverConfiguration, SolverSettings} from "../SolverSettings/solverSettings";
import {SolverInput} from "../SolverInput/solverInput";
import {SolverSpinner} from "../SolverSpinner/solverSpinner";
import {SolverLog} from "../SolverLog/solverLog";
import {Alert, Modal, Button} from 'antd';
import { ExclamationCircleOutlined } from '@ant-design/icons';


if (!WebAssembly.instantiateStreaming) { // polyfill
    WebAssembly.instantiateStreaming = async (resp, importObject) => {
        const source = await (await resp).arrayBuffer();
        return await WebAssembly.instantiate(source, importObject);
    };
}

const go = new (window as any).Go();
go.argv = [''];
let mod: any, inst: any;
WebAssembly.instantiateStreaming(fetch("sat.wasm", {cache: 'no-cache'}), go.importObject).then((result) => {
    mod = result.module;
    inst = result.instance;
}).catch((err) => {
    console.error(err);
});

let wasInited: boolean = false;

export class EventCollector {
    trace(text: string) {
        console.debug(text);
    }

    startProcessing(text: string) {
        console.warn(text);
    }

    endProcessing(text: string) {
        console.warn(text);
    }

    startNewRun() {

    }
}

let solve: Optional<(
    input: string,
    eventCollector: EventCollector,
    callback: (err: string | undefined, result: number | undefined, assignment: any) => void,
    loaderName: string,
    solverName: string
) => void> = null;
async function run() {
    await go.run(inst);
    inst = await WebAssembly.instantiate(mod, go.importObject);
}

async function init() {
    if (wasInited) return;
    wasInited = true;
    run();
    solve = (window as any).solve;
}

type Optional<F> = F | null;

type SatResult = { sat: number, assignment: any };

const performSolve = async (input: string, eventCollector: EventCollector, configuration: SolverConfiguration) => {
    return new Promise<SatResult>(((resolve, reject) => {
        if (solve) {
            setTimeout(() => {
                solve!(input, eventCollector, (err, result, assignment) => {
                    if (err !== undefined) {
                        reject(err);
                    } else if (result !== undefined) {
                        resolve({
                            sat: result,
                            assignment,
                        });
                    } else {
                        reject("Unknown error.");
                    }
                }, configuration.loaderName, configuration.solverName);
            }, 0);
        } else {
            reject("Solver WASM was loaded incorrectly cannot find the method to call.");
        }
    }))
};

export const Solver: React.FunctionComponent<{}> = () => {

    const [codeContent, setCodeContent] = React.useState<string>(`And (Var "a") (Or (Not (Var "a")) F)`);
    const [defaultConfig, setDefaultConfig] = React.useState<SolverConfiguration>(INITIAL_SETTINGS);
    const [isRunning, setIsRunning] = React.useState<boolean>(false);
    const [result, setResult] = React.useState<SatResult | undefined | string>(undefined);
    const eventCollector = React.useRef<EventCollector>(new EventCollector());

    setTimeout(init, 1000);

    const onSolve = async (conf: SolverConfiguration) => {
        setIsRunning(true);
        eventCollector.current.startNewRun();
        setDefaultConfig(conf);
        setTimeout(async () => {
            try {
                console.warn("conf", conf);
                const result = await performSolve(codeContent, eventCollector.current, conf);
                setResult(result);
                setIsRunning(false);
            } catch (e) {
                setResult(e.toString());
                setIsRunning(false);
            }
        }, 0);
    };

    if (isRunning) {
        return (
            <>
                <SolverLog
                    onUpdateCollector={(collector) => {console.log("UPD"); eventCollector.current = collector;}}
                />
                <SolverSpinner />
            </>
        );
    }

    const onShowAssignment = () => {
        if (typeof result === 'string' || result === undefined) {
            return;
        }
        Modal.confirm({
            icon: <ExclamationCircleOutlined />,
            content: <div>
                <div>
                    This is assignment of variables that satisfies the given input formula.
                    Please note that the variables that can have any value are omitted from the result.
                </div>
                <div>
                    {JSON.stringify(result.assignment, null, 2)}
                </div>
            </div>,
            onOk() {
                console.log('OK');
            },
            onCancel() {
                console.log('Cancel');
            },
        });
    };

    let resultNode = null;
    if (result === undefined) {
        // Do nothing
    } else if (typeof result === 'string') {
        resultNode = (
            <div style={{marginBottom: '20px',}}>
                <Alert
                    message="Error"
                    description={result}
                    type="error"
                    showIcon
                />
            </div>
        );
    } else {
        resultNode = (
            <div style={{marginBottom: '20px',}}>
                <Alert
                    message="Formula solved"
                    description={<div>
                        <div style={{marginBottom: '10px',}}>
                            {`Result: ${result.sat === 1 ? 'SAT': 'UNSAT'}`}
                        </div>
                        {result.sat === 1 ? (<Button onClick={onShowAssignment}>Show assignment</Button>) : null}
                    </div>}
                    type={result.sat === 0 ? "error" : "success"}
                    showIcon
                />
            </div>
        );
    }

    return (
        <div style={{ textAlign: 'left', }}>
            <div style={{ position: 'relative', left: '-90px', }}>
                <SolverSettings
                    initialValues={defaultConfig}
                    onRequestSolve={onSolve}
                />
            </div>
            {resultNode}
            <SolverInput
                input={codeContent}
                onInputChange={setCodeContent}
            />
        </div>
    );
};
