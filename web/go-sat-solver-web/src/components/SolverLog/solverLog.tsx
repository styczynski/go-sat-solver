import React from "react";
import ReactDOM from "react-dom";
import {EventCollector} from "../Solver/solver";
import { LogInner } from "./logInner";

export interface Props {
    onUpdateCollector: (newCollector: EventCollector) => void;
}

export class SolverLog extends React.Component<Props, {}> {
    
    collector: EventCollector;
    logContents: Array<string> = [];
    contentRef: HTMLDivElement | null = null;

    constructor(props: Props) {
        super(props);
        this.collector = new EventCollector();
        this.updateContent = this.updateContent.bind(this);
    }
    
    updateContent() {
        setTimeout(() => {
            ReactDOM.render(<LogInner content={this.logContents}/>, this.contentRef!)
        }, 0);
    };

    shouldComponentUpdate() {
        return false;
    }
    
    componentDidMount() {
        const self = this;
        this.collector = new (class extends EventCollector {
            trace(text: string) {
                self.logContents = [...self.logContents, text];
                self.updateContent();
            }

            startProcessing(text: string) {
                self.logContents = [...self.logContents, text];
                self.updateContent();
            }

            endProcessing(text: string) {
                self.logContents = [...self.logContents, text];
                self.updateContent();
            }

            startNewRun() {
                self.logContents = [];
                self.updateContent();
            }
        })();
        this.props.onUpdateCollector(this.collector);
    }
    
    render() {
        this.props.onUpdateCollector(this.collector);
        return (
            <div style={{color: 'white'}}>
                <div ref={(el) => {console.error("SETREF", el); if (el) this.contentRef = el;}} />
            </div>
        );
    }
};
