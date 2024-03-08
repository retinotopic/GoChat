import { createSignal, Switch,Match } from "solid-js";

function chat() {
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
    <div></div>
  );

}