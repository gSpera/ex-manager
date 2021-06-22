class Exploit extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            name: props.name,
            flags: props.flags,
            running: props.exploitState,
        };
    }

    componentDidMount() {
        this.timer = setInterval(
            () => this.setState({ name: this.state.name + ":" }),
            10000
        );
    }
    componentDidUnmount() {
        clearInterval(this.timer);
    }

    render() {
        return <div className="exploit box">
            <div className="status-line level">
                <span className="name title is-4">{this.state.name}</span>
                {
                    this.state.exploitState == "Running"
                        ? <button onClick={() => this.setState({ exploitState: "Paused" })} className="button is-danger">Stop</button>
                        : <button onClick={() => this.setState({ exploitState: "Running" })} className="button is-success">Run</button>
                }
            </div>
            <details>
                <summary className="title is-6 is-clickable m-1">Flags</summary>
                <table className="target-info table is-hoverable is-fullwidth">
                    <thead>
                        <tr>
                            <th>Target</th>
                            <th>Flags Taken</th>
                            <th>Fixed</th>
                        </tr>
                    </thead>

                    <tbody>
                        {
                            this.state.flags.map(
                                (target) =>
                                    <tr key={target.target}>
                                        <td key={target.target + "-target"}>{target.target}</td>
                                        <td key={target.target + "-address"}>{target.flags}</td>
                                        <td key={target.target + "-flags"}>{target.fixed ? 'X' : '-'}</td>
                                    </tr>
                            )
                        }
                    </tbody>
                </table>
            </details>
        </div >;
    }
}
class Component extends React.Component {
    constructor(props) { super(props); }

    render() {
        return <h1>A</h1>;
    }
}

let container = document.querySelector("#add-service");
ReactDOM.render(
    <Exploit name="Exploit" flags={[
        { "target": "0", "flags": 1, "fixed": true },
        { "target": "1", "flags": 7, "fixed": true },
        { "target": "7", "flags": 1, "fixed": false }

    ]} exploitState="Running" />,
    container,
);