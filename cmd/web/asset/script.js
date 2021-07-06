class Services extends React.Component {
    constructor(props) {
        super(props);
        this.state = { services: [] };
        this.update = this.update.bind(this);
    }

    componentDidMount() {
        const timer = setInterval(() => this.update(), 1000);
        this.setState({ ...this.state, timer });
    }

    componentWillUnmount() {
        clearInterval(this.state.timer);
    }

    update() {
        fetch("/api/sessionStatus")
            .then(r => r.json())
            .then(r => this.setState({ ...this.state, services: r["Services"] }))
            .catch(err => console.error(err));
    }

    render() {
        return <React.Fragment>
            {
                this.state.services.map(service => <Service key={service} name={service} refUploadExploit={this.props.modalRef} />)
            }
        </React.Fragment>
    }
}

class Service extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            exploits: [],
        };
        this.update = this.update.bind(this);
        this.newExploit = this.newExploit.bind(this);
        this.update();
    }

    newExploit() {
        const modal = this.props.refUploadExploit.current;
        modal.setState({
            ...modal.state,
            service: this.props.name,
            hidden: false,
        });
    }

    componentDidMount() {
        const timer = setInterval(() => this.update(), 1000);
        this.setState({ ...this.state, timer });
    }

    componentWillUnmount() {
        clearInterval(this.state.timer);
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
            <div className="level">
                <h2 className="title is-2">{this.props.name}</h2>
                <button onClick={this.newExploit} className="icon is-primary is-large is-inverted button">+</button>
            </div>
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

class GlobalModal extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            hidden: true,
            exploit: "",
            cmd: "",
        };
        this.fileInput = React.createRef();
        this.uploadFile = this.uploadFile.bind(this);
    }

    uploadFile(e) {
        // csrf??
        e.preventDefault();
        const content = new FormData();
        content.append("service", this.state.service);
        content.append("exploit", this.state.exploit);
        content.append("cmd", this.state.cmd);
        content.append("file", this.fileInput.current.files[0]);

        fetch("/api/uploadExploit", {
            method: "POST",
            body: content,
        })
            .then(r => r.json())
            .then(r => {
                if (r["Ok"]) {
                    this.setState({
                        ...this.state,
                        service: "ERROR",
                        hidden: true,
                    });
                    // notification
                } else {
                    alert("Cannot upload file");
                    console.log(r["Reason"]);
                    // notification
                }
            })
            .catch(err => console.error(err));
    }

    onChange(what, e) {
        let newState = this.state;
        switch (what) {
            case "name":
                newState.exploit = e.target.value;
                break;
            case "cmd":
                newState.cmd = e.target.value;
                break;
            default:
                alert("what");
                break;
        }

        this.setState(newState);
    }

    render() {
        return <div className={"modal " + (this.state.hidden ? "" : "is-active")}>
            <div className="modal-background"></div>
            <div className="modal-card">
                <header className="modal-card-head">
                    <div className="modal-card-title">
                        <h1 className="title is-3">Upload exploit</h1>
                        <h2 className="subtitle is-4">Service: {this.state.service}</h2>
                    </div>
                    <button className="delete" aria-label="close" onClick={() => this.setState({ ...this.state, hidden: true })}></button>
                </header>
                <div className="modal-card-body">
                    <form id="modal-upload" onSubmit={this.uploadFile}>
                        <input placeholder="Exploit Name" className="input" type="text" name="name" onChange={(e) => this.onChange("name", e)} />
                        <input placeholder="Exploit Command" className="input" type="text" name="cmd" onChange={(e) => this.onChange("cmd", e)} />
                        <input className="input" type="file" name="file" ref={this.fileInput} />
                        <span className="notification is-warning is-small">The file will be overwritten</span>
                        <button className="button" onClick={this.uploadFile}>Upload</button>
                    </form>
                </div>
            </div>
        </div >;
    }
}

class ServiceAddComponent extends React.Component {
    constructor(props) { super(props); this.button = this.button.bind(this); }

    button() {
        fetch("/api/newService?name=" + this.state.content)
            .then(r => r.json())
            .then(r => {
                if (r["Ok"]) {
                    // notification
                } else {
                    // notification
                    alert("Error");
                }
            })
            .catch(err => console.error(err))
    }
    render() {
        return <React.Fragment>
            <div className="title is-2 has-text-centered">
                Add new service
            </div>
            <input type="text" className="input mgh-large" name="service-name" placeholder="Service name" onChange={(e) => this.setState({ ...this.state, content: e.target.value })} />
            <button id="add-service-btn" className="is-large button mgh-large" onClick={this.button}>Add Service</button>
        </React.Fragment>;
    }
}


class StackedBarGraph extends React.Component {
    constructor(props) {
        super(props);
        const ref = React.createRef();

        this.state = {
            flags: { success: [], expired: [] },
            ref,
        };

        this.update = this.update.bind(this);
    }

    componentDidMount() {
        const timer = setInterval(() => this.update(), 1000);
        const graph = new Chart(this.state.ref.current, {
            title: "Submitted Flags",
            type: "bar",
            options: { dataColors: ['hsl(171, 100%, 41%)', 'hsl(348, 100%, 61%)'], scales: { x: { stacked: true }, y: { stacked: true } }, responsive: false },
            xLabel: "Time (Batch = 10s)",
            yLabel: "Flags",
            data: {
                labels: ["9", "8", "7", "6", "5", "4", "3", "2", "1", "Now"],
                datasets: [
                    {
                        label: "Success",
                        data: this.state.flags.success,
                        backgroundColor: ['rgba(75, 192, 192, 0.7)'],
                    },
                    {
                        label: "Expired",
                        data: this.state.flags.expired,
                        backgroundColor: ['rgba(255, 159, 64, 0.7)'],
                    },
                    {
                        label: "Invalid",
                        data: this.state.flags.success,
                        backgroundColor: ['rgba(255, 159, 64, 1)'],
                    },
                    {
                        label: "Not Submitted",
                        data: this.state.flags.success,
                        backgroundColor: ['rgba(201, 203, 207, 0.7)'],
                    },
                    {
                        label: "Already Submitted",
                        data: this.state.flags.success,
                        backgroundColor: ['rgba(255, 205, 86, 0.2)'],
                    },
                    {
                        label: "Own",
                        data: this.state.flags.expired,
                        backgroundColor: ['rgba(153, 102, 255, 0.2)'],
                    },
                    {
                        label: "Nop",
                        data: this.state.flags.expired,
                        backgroundColor: ['rgba(255, 159, 64, 0.7)'],
                    },
                    {
                        label: "Offline",
                        data: this.state.flags.expired,
                        backgroundColor: ['rgba(54, 162, 235, 0.2)'],
                    },
                    {
                        label: "Offline Service",
                        data: this.state.flags.expired,
                        backgroundColor: ['rgba(255, 159, 64, 0.7)'],
                    },
                ]
            }
        })
        this.setState({
            ...this.state,
            graph,
            timer,
        });
    }
    componentWillUnmount() {
        clearInterval(this.state.timer)
    }
    update() {
        fetch("/api/submitterStatus")
            .then(r => r.json())
            .then(r => {
                const filter = requiredState => r.map(batch => batch.filter(flag => flag["Status"] == requiredState));
                const process = requiredState => filter(requiredState).map(batch => batch.length).reverse();
                const flagsSuccess = process("SUCCESS");
                const flagsExpired = process("EXPIRED");
                const flagsNotSubmitted = process("NOT-SUBMITTED");
                const flagsInvalid = process("INVALID");
                const flagsAlready = process("ALREADY-SUBMITTED");
                const flagsOwn = process("TEAM-OWN");
                const flagsNop = process("TEAM-NOP");
                const flagsOffline = process("OFFLINE-CTF");
                const flagsServiceOffline = process("OFFLINE-SERVICE");

                let graph = this.state.graph;
                graph.data.datasets[0].data = flagsSuccess;
                graph.data.datasets[1].data = flagsExpired;
                graph.data.datasets[2].data = flagsNotSubmitted;
                graph.data.datasets[3].data = flagsInvalid;
                graph.data.datasets[4].data = flagsAlready;
                graph.data.datasets[5].data = flagsOwn;
                graph.data.datasets[6].data = flagsNop;
                graph.data.datasets[7].data = flagsOffline;
                graph.data.datasets[8].data = flagsServiceOffline;
                graph.update();

                this.setState({
                    ...this.state,
                    graph,
                    flags: {
                        success: flagsSuccess,
                        expired: flagsExpired,
                        notSubmitted: flagsNotSubmitted,
                        invalid: flagsInvalid,
                        already: flagsAlready,
                        own: flagsOwn,
                        nop: flagsNop,
                        offline: flagsOffline,
                        serviceOffline: flagsServiceOffline,
                    },
                });
            })
            .catch(err => console.error(err));
    }

    render() {
        return <canvas ref={this.state.ref}></canvas>;
    }
}
const modalRef = React.createRef();
const serviceRef = React.createRef();
ReactDOM.render(
    <GlobalModal ref={modalRef} />,
    document.getElementById("global-modal"),
)
fetch("/api/sessionStatus")
    .then(r => r.json())
    .then(r => {
        document.querySelector("#navbar > h1").innerText = r["Name"];

        ReactDOM.render(
            <Services services={r["Services"]} modalRef={modalRef} ref={serviceRef} />,
            document.getElementById("services-root")
        );
    })
    .catch(err => console.error(err));

ReactDOM.render(
    <ServiceAddComponent serviceRef={serviceRef} />,
    document.getElementById("add-service"),
);

ReactDOM.render(
    <StackedBarGraph />,
    document.getElementById("graph-stacked-root"),
);