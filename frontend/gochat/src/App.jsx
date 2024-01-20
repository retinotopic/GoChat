import { createSignal, Switch,Match } from "solid-js";
import login from '.components/login';
import chat from '.components/chat';
function App() {
  const [httpcode, setHttpCode] = createSignal(418);
  async function trytologin() {
    const response = await fetch("http://localhost:8080/chat",{
      method: 'GET',
      credentials: 'include'
    });
    const resp = await response.json();
    setHttpCode(resp.status);
  }
  return (
    <Switch fallback={<p>loading...</p>}>
      <Match when={httpcode()===401} >
        <login/>
      </Match>
      <Match when={httpcode()!=401}> 
        chat app will be here
      </Match>
    </Switch>
  );

}

export default App;
