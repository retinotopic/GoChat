function login() {
    return (
      <div>
        <div>
          <a href="http://localhost:8080/google/BeginLoginCreate">SIGN IN WITH GOOGLE</a>
        </div>
        <div>
          <form action="http://localhost:8080/gfirebase/BeginLoginCreate">
            <input type="email" name="email"/>
            <button type="submit">SIGN IN WITH EMAIL</button>
          </form>
        </div>
      </div>
    );
  }