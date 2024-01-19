import { render } from 'solid-js/web';
import { Router, Route, Routes } from "@solidjs/router";
import App from './App';

render(
  () => <App />, document.getElementById("root")
);
