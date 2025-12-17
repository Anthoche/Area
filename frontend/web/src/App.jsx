import {BrowserRouter, Route, Routes} from "react-router-dom";
import Login from "./components/Login/Login";
import Register from "./components/Register/Register";
import Homepage from "./components/Homepage/Homepage";
import CreateAcc from "./components/CreateAcc/CreateAcc";
import WelcomePage from "./components/WelcomePage/WelcomePage.jsx";


function App() {
    return (
        <BrowserRouter>
            <Routes>
                <Route path="/" element={<WelcomePage/>}/>
                <Route path="/home" element={<Homepage/>}/>
                <Route path="/login" element={<Login/>}/>
                <Route path="/register" element={<Register/>}/>
                <Route path="/createacc" element={<CreateAcc/>}/>
            </Routes>
        </BrowserRouter>
    );
}

export default App;
