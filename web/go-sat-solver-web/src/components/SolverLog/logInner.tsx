import React, {useEffect} from "react";
import "./log.less";

export interface Props {
    content: Array<string>;
}

export const LogInner: React.FunctionComponent<Props> = ({ content }) => {
    return (
        <div className="Log">
            {content.map(line => {
                if (line.indexOf("|") > -1) {
                    let tokens = line.split("|");
                    tokens = tokens.splice(1);
                    line = tokens.join("|");
                } else {
                    line = line.replace("[0]", "").replace("[init]", "");
                }
                return (
                    <div>
                        {line.split('\n').map(l => (<div>{l}</div>))}
                    </div>
                );
            })}
        </div>
    );
};
