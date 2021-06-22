class Component extends React.Component {
    constructor(props) { super(props); }

    render() {
        return <h1>A</h1>;
    }
}

container = document.querySelector("#add-service");
ReactDOM.render(
    <Component></Component>,
    container,
);