import { useState, useEffect } from 'react'
import {GreetService} from "../bindings/wails-react-3-demo";
import {Events, WML} from "@wailsio/runtime";

function App() {
  const [name, setName] = useState('');
  const [result, setResult] = useState('点击尝试应用更新 👇');
  const [time, setTime] = useState('Listening for Time event...');

  const doGreet = () => {
    let localName = name;
    if (!localName) {
      localName = 'anonymous';
    }
    GreetService.Greet(localName).then((resultValue) => {
      setResult(resultValue);
    }).catch((err) => {
      console.log(err);
    });
  }

    const doGetPath = () => {

        GreetService.GetCurrentPath().then((resultValue) => {
            setResult(resultValue);
        }).catch((err) => {
            console.log(err);
        });
    }

  useEffect(() => {
    Events.On('time', (timeValue) => {
      setTime(timeValue.data);
    });
    // Reload WML so it picks up the wml tags
    WML.Reload();
  }, []);

  return (
    <div className="container">
      <div>
        {/*<a wml-openURL="https://wails.io">*/}
        {/*  <img src="/wails.png" className="logo" alt="Wails logo"/>*/}
        {/*</a>*/}
        {/*<a wml-openURL="https://reactjs.org">*/}
        {/*  <img src='/react.svg' className="logo react" alt="React logo"/>*/}
        {/*</a>*/}
      </div>
      <h1>Wails + React</h1>
      <div className="result">{result}</div>
      <div className="card">
        <div className="input-box">
          <input className="input" value={name} onChange={(e) => setName(e.target.value)} type="text" autoComplete="off"/>
          <button className="btn" onClick={doGreet}>Greet</button>
          <button className="btn" onClick={doGetPath}>路径</button>
        </div>
        {/*  <div>这个是更新后的版本</div>*/}
      </div>
      <div className="footer">
        {/*<div><p>Click on the Wails logo to learn more</p></div>*/}
        <div><p>{time}</p></div>
      </div>
    </div>
  )
}

export default App
