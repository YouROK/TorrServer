package template

func (t *Template) parseMainPage() {
	t.parsePage("mainPage", mainPage)
}

const mainPage = `
<!DOCTYPE html>
<html>
<head>
    <title>TorrServer</title>

    <meta charset="utf-8"/>
    <meta name="viewport" content="minimum-scale=1, initial-scale=1, width=device-width"/>
    <script src="https://unpkg.com/react@latest/umd/react.development.js" crossorigin="anonymous"></script>
    <script src="https://unpkg.com/react-dom@latest/umd/react-dom.development.js"></script>
    <script src="https://unpkg.com/@material-ui/core@latest/umd/material-ui.development.js" crossorigin="anonymous"></script>
    <script src="https://unpkg.com/babel-standalone@latest/babel.min.js" crossorigin="anonymous"></script>
    <!-- Fonts to support Material Design -->
    <link rel="stylesheet" href="https://fonts.googleapis.com/css?family=Roboto:300,400,500,700&display=swap"/>
    <!-- Icons to support Material Design -->
    <link rel="stylesheet" href="https://fonts.googleapis.com/icon?family=Material+Icons"/>

</head>
<body>
<div id="root"></div>

<script type="text/babel">

    class App extends React.Component {
        constructor(props) {
            super(props);
            this.state = {
                error: null,
                version: "",
            };
        }

        componentDidMount() {
            fetch("/echo")
                .then(res => res.text())
                .then(
                    (txt) => {
                        this.setState({
                            version: txt
                        });
                    }
                );
        }

        componentDidUpdate() {
            document.title = 'TorrServer ' + this.state.version;
        }

        render() {
            return (
                <div>
                    <Torrent/>
                </div>
            );
        }
    }

    class TorrentList extends React.Component {
		constructor(props) {
            super(props);
            this.state = {date: new Date()};
        }
        componentDidMount() {
            this.timerID = setInterval(
                () => this.tick(),
                1000
            );
        }

        componentWillUnmount() {
            clearInterval(this.timerID);
        }

        tick() {
            this.setState({
                date: new Date()
            });
        }

        render() {
            return (
                <div>
                    <h1>Привет, мир!</h1>
                    <h2>Сейчас {this.state.date.toLocaleTimeString()}.</h2>
                </div>
            );
        }
    }


    const {
        colors,
        CssBaseline,
        ThemeProvider,
        Typography,
        Container,
        makeStyles,
        createMuiTheme,
        Box,
        SvgIcon,
        Link,
    } = MaterialUI;

    // Create a theme instance.
    const theme = createMuiTheme({
        palette: {
            primary: {
                main: '#556cd6',
            },
            secondary: {
                main: '#19857b',
            },
            error: {
                main: colors.red.A400,
            },
            background: {
                default: '#fff',
            },
        },
    });

    function LightBulbIcon(props) {
        return (
            <SvgIcon {...props}>
                <path d="M9 21c0 .55.45 1 1 1h4c.55 0 1-.45 1-1v-1H9v1zm3-19C8.14 2 5 5.14 5 9c0 2.38 1.19 4.47 3 5.74V17c0 .55.45 1 1 1h6c.55 0 1-.45 1-1v-2.26c1.81-1.27 3-3.36 3-5.74 0-3.86-3.14-7-7-7zm2.85 11.1l-.85.6V16h-4v-2.3l-.85-.6C7.8 12.16 7 10.63 7 9c0-2.76 2.24-5 5-5s5 2.24 5 5c0 1.63-.8 3.16-2.15 4.1z"/>
            </SvgIcon>
        );
    }

    const useStyles = makeStyles(theme => ({
        root: {
            margin: theme.spacing(6, 0, 3),
        },
        lightBulb: {
            verticalAlign: 'middle',
            marginRight: theme.spacing(1),
        },
    }));

    ReactDOM.render(
        <App/>,
        document.getElementById('root')
    );
</script>
</body>
</html>
`
