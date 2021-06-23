function Services(props) {
    return <div>
        {
            props.services.map(service => <Service key={service} name={service} />)
        }
    </div>
}
class Service extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            exploits: [],
        };
        this.update = this.update.bind(this);
        this.update();
    }

    componentDidMount() {
        const timer = setInterval(() => this.update(), 1000);
        this.setState({ ...this.state, timer });
    }

    update() {
        fetch("/api/serviceStatus?service=" + this.props.name)
            .then(r => r.json())
            .then(r => this.setState({
                ...this.state,
                exploits: r["Exploits"],
            }))
            .catch(err => console.error(err));
    }

    render() {
        return <div className="service box">
            <h2 className="title is-2">{this.props.name}</h2>
            {
                this.state.exploits.map(exploit =>
                    <Exploit service={this.props.name} name={exploit} key={this.props.name + "-" + exploit} />
                )
            }
        </div>;
    }
}

class Exploit extends React.Component {
    constructor(props) {
        super(props);

        this.state = {
            targets: [],
        };
        this.update = this.update.bind(this);
        this.update();
    }

    update() {
        fetch("/api/exploitStatus?service=" + this.props.service + "&exploit=" + this.props.name)
            .then(r => r.json())
            .then(r => this.setState({
                ...this.state,
                targets: r["Targets"],
                running: r["State"],
            }))
            .catch(err => console.error(err));
    }
    componentDidMount() {
        const timer = setInterval(
            () => this.update(),
            1000,
        );
        this.setState({ ...this.state, timer });
    }

    componentWillUnmount() {
        clearInterval(this.state.timer);
    }

    changeState(state) {
        fetch("/api/exploitChangeState?service=" + this.props.service + "&exploit=" + this.props.name + "&state=" + state)
            .then(r => r.json())
            .then(r => {
                if (r["ok"]) {
                    this.setState({
                        ...this.state,
                        running: state,
                    })
                } else {
                    console.error(JSON.stringify(r))
                    alert("Cannot change state to:" + state);
                }
            })
            .catch(err => console.error(err));
    }

    render() {
        return <div className="exploit box">
            <div className="status-line level">
                <span className="name title is-4">{this.props.name}</span>
                {
                    this.state.running == "Running"
                        ? <button onClick={() => this.changeState("Paused")} className="button is-success is-outlined">Running</button>
                        : <button onClick={() => this.changeState("Running")} className="button is-danger">Paused</button>
                }
            </div>
            <details>
                <summary className="title is-6 is-clickable m-1">
                    Flags: <span>{this.state.targets.reduce((res, add) => res + add.Flags.length, 0)}</span>
                </summary>

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
                            this.state.targets.map(
                                target =>
                                    <tr key={target.Name}>
                                        <td key={target.Name + "-target"}>{target.Name}</td>
                                        <td key={target.Name + "-address"}>{target.Flags.length}</td>
                                        <td key={target.Name + "-flags"}>{target.Fixed ? 'X' : '-'}</td>
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

fetch("/api/sessionStatus")
    .then(r => r.json())
    .then(r => {
        document.querySelector("#navbar > h1").innerText = r["Name"];

        ReactDOM.render(
            <Services services={r["Services"]} />,
            document.getElementById("services-root")
        );
    })
    .catch(err => console.error(err));
