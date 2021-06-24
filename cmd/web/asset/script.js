function Services(props) {
    return <div>
        {
            props.services.map(service => <Service key={service} name={service} refUploadExploit={props.modalRef} />)
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

const modalRef = React.createRef();
ReactDOM.render(
    <GlobalModal ref={modalRef} />,
    document.getElementById("global-modal"),
)
fetch("/api/sessionStatus")
    .then(r => r.json())
    .then(r => {
        document.querySelector("#navbar > h1").innerText = r["Name"];

        ReactDOM.render(
            <Services services={r["Services"]} modalRef={modalRef} />,
            document.getElementById("services-root")
        );
    })
    .catch(err => console.error(err));
