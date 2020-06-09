import React from "react";
import { Spin } from 'antd';
import { LoadingOutlined } from '@ant-design/icons';

export const SolverSpinner = () => {
    return (
        <Spin indicator={<LoadingOutlined style={{ fontSize: 24 }} spin />} />
    );
};
