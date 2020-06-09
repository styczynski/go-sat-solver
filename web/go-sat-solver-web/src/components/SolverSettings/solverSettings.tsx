import React from 'react';
import { Select, Radio, Form, Input, Button } from 'antd';

const layout = {
    labelCol: { span: 8 },
    wrapperCol: { span: 16 },
};
const tailLayout = {
    wrapperCol: { offset: 8, span: 16 },
};

const AVAILABLE_SOLVERS = [ "cdcl", "naive" ];
const AVAILABLE_INPUTS = [ "haskell", "cnf" ];

export interface SolverConfiguration {
    solverName: string;
    loaderName: string;
}

export const INITIAL_SETTINGS: SolverConfiguration = {
    solverName: "cdcl",
    loaderName: "haskell",
};

export const SolverSettings: React.FunctionComponent<{
    onRequestSolve?: (config: SolverConfiguration) => void;
    initialValues: SolverConfiguration,
}> = ({ onRequestSolve, initialValues, }) => {
    const [form] = Form.useForm();

    const onFinish = (values: any) => {
        if (onRequestSolve) {
            onRequestSolve(values);
        }
    };

    const onReset = () => {
        form.resetFields();
    };

    return (
        <Form {...layout} initialValues={initialValues || INITIAL_SETTINGS} form={form} name="control-hooks" onFinish={onFinish}>
            <Form.Item name="solverName" label="Solver used" rules={[{ required: true }]}>
                <Select
                    placeholder="Select a solver to be used"
                    allowClear
                >
                    {AVAILABLE_SOLVERS.map(solverName => (<Select.Option value={solverName}>{solverName}</Select.Option>))}
                </Select>
            </Form.Item>
            <Form.Item name="loaderName" label="Input format" rules={[{ required: true }]}>
                <Select
                    placeholder="Select an input format"
                    allowClear
                >
                    {AVAILABLE_INPUTS.map(loaderName => (<Select.Option value={loaderName}>{loaderName}</Select.Option>))}
                </Select>
            </Form.Item>
            <Form.Item {...tailLayout}>
                <Button type="primary" htmlType="submit">
                    Solve
                </Button>
                <Button htmlType="button" onClick={onReset}>
                    Reset
                </Button>
            </Form.Item>
        </Form>
    );
};
