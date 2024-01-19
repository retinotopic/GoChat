function login() {
    async function authentication_request() {
      const response = await fetch("http://localhost:8080/login",{
        method: 'GET',
        headers: {
          'Authorization': 'Bearer ' + localStorage.getItem('jwt')
        }
      });
      const resp = await response.json();
      setHttpCode(resp.status);
    }
    return (
      <div>
        <button onClick={authentication_request}>LOGIN</button>
      </div>
    );
  }