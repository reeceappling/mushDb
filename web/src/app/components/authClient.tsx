'use client'

import React, {useEffect, useState} from "react";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {CredentialResponse, GoogleLogin} from "@react-oauth/google";
import {ReadonlyURLSearchParams, useSearchParams} from "next/navigation";

// TODO: ADD LOGOUT BUTTON SOMEWHERE!
export interface AuthAreaProps {
    successUrl: string,
    loggedIn: boolean
}

export default function AuthArea( // TODO: any depth?????
    {
        successUrl,
        loggedIn
    }: AuthAreaProps) {
    // const [cookies, setCookie, removeCookie] = useCookies(['SessionId']); // TODO: may be needed
    const [user, setUser] = useState<string>("")
    const [pass, setPass] = useState<string>("")
    const [remember, setRemember] = useState<boolean>(false)
    const [err, setErr] = useState<string | undefined>()
    useEffect(() => {
        if (loggedIn) {
            location.assign(successUrl) // see redirectToBasePage
            return
        }
    }, [loggedIn])
    const setSessionCookie = (sessId: string) => {
        // TODO: IS PATH OK?
        //setCookie('SessionId', sessId, { path: '/', expires: new Date(Date.now() + (3600000*2)) }); // Expires in 1 hour
        location.assign(successUrl) // see redirectToBasePage
    }
    const handleGoogleLoginSuccess = (sessId: string) => {
        console.log("response creds:",sessId) // TODO: FIX
        // TODO: https://developers.google.com/identity/protocols/oauth2/javascript-implicit-flow
        // TODO: redirectUri
        // TODO: make sure has scopes https://www.googleapis.com/auth/userinfo.email and https://www.googleapis.com/auth/userinfo.profile
        // TODO: DO THESE: https://developers.google.com/identity/protocols/oauth2
        // TODO: creds response is id token?
        // TODO: CREATE SESSION
        // TODO: REDIRECT
    }
    const handleLoginSuccess = (sessId: string)=>{
        // setCookie('SessionId', sessId, { // TODO: may be needed
        //     path: '/', // Makes the cookie available throughout the entire site
        //     maxAge: 3600, // Cookie expires in 1 hour (in seconds)
        //     // other options like 'secure', 'httpOnly', 'sameSite' can be added
        // })
        // TODO: redirect to correct area // see redirectToBasePage
    }
    return (<>
        {/*// <div className={"fullPage"}>*/}
            {/* TODO: ERROR FOR FAILED LOGIN */}
            <div className={"fixCenterScreen"}></div>
            <div className={"centerH"}>
                <ErrorDisplay err={err}/>{/* TODO: headerLevel ok? */}
            </div>
            <div className="centerH">
                {/* GOOGLE SIGN IN/UP*/}
                <SignInArea onLogin={handleGoogleLoginSuccess}/>
            </div>
            {/*<div className="loginRow">*/}
            {/*    <a href={"/guestLogin"}>{"Continue as guest"}</a>*/}
            {/*</div>*/}
        {/*// </div>*/}
        </>
    )
}

// export function SignupArea( // TODO: DO THIS WHOLE THING
//     {
//         successUrl,
//         loggedIn
//     }: AuthAreaProps) {
//     //const [cookies, setCookie, removeCookie] = useCookies(['SessionId']);
//     const [user, setUser] = useState<string | undefined>()
//     const [pass, setPass] = useState<string | undefined>()
//     const [email, setEmail] = useState<string | undefined>()
//     const [adminUser, setAdminUser] = useState<string | undefined>()
//     const [adminPass, setAdminPass] = useState<string | undefined>()
//     const [remember, setRemember] = useState<boolean>(false)
//     const [err, setErr] = useState<string | undefined>()
//     useEffect(() => {
//         if (loggedIn) {
//             router.replace(successUrl)
//             return
//         }
//     }, [loggedIn])
//     const trySignupUserPass = () => {
//         if(!pass||!user){
//             setErr("username or password must exist")
//             return
//         }
//         const hashedPassword = bcrypt.hashSync(pass)
//         fetch(BaseExternalUrl+'/signup', {
//             method: 'POST',
//             headers: clientPostRequestHeaders,
//             body: JSON.stringify({
//                 username: user,
//                 password: hashedPassword,
//             }),
//         })
//             .then((response)=>{
//                 if (!response.ok){
//                     throw new Error(`Response status: ${response.status}`);
//                 }
//                 if (response.status!==202){
//                     throw new Error(`Status code was not 202, it was ${response.status}`)
//                 }
//                 return response.text()
//             })
//             .then(sessId=>{
//                 // TODO: SET SESSION ID AS COOKIE?
//                 // TODO: IS PATH OK?
//                 setCookie('SessionId', sessId, { path: '/', expires: new Date(Date.now() + (3600000*2)) }); // Expires in 1 hour
//             })
//             .catch((e)=>{setErr(e)})
//     }
//     const trySignupViaAdmin = () => {
//         if(!pass||!user){
//             setErr("username and password must exist")
//             return
//         }
//         if(!adminPass||!adminUser){
//             setErr("admin username and password must exist")
//             return
//         }
//         const hashedAdminPassword = bcrypt.hashSync(adminPass)
//         const hashedPassword = bcrypt.hashSync(pass)
//         fetch(BaseExternalUrl+'/signup', {
//             method: 'POST',
//             headers: {
//                 'isAdmin': "true", // TODO: delete or no?
//                 //Accept: 'application/json', // TODO: text
//                 'Content-Type': 'application/json',
//             },
//             body: JSON.stringify({
//                 username: user,
//                 password: hashedPassword,
//             }),
//         })
//             .then((response)=>{
//                 if (!response.ok){
//                     throw new Error(`Response status: ${response.status}`);
//                 }
//                 if (response.status!==202){
//                     throw new Error(`Status code was not 202, it was ${response.status}`)
//                 }
//                 return response.text()
//             })
//             .then(sessId=>{
//                 // TODO: SET SESSION ID AS COOKIE?
//                 // TODO: IS PATH OK?
//                 setCookie('SessionId', sessId, { path: '/', expires: new Date(Date.now() + (3600000*2)) }); // Expires in 1 hour
//             })
//             .catch((e)=>{setErr(e)})
//     }
//     const handleGoogleCreateSuccess = (credentialResponse: CredentialResponse) => {
//         if (!credentialResponse.credential) {
//             setErr("No credential token found on google oAuth response")
//             return
//         }
//         // TODO: THIS
//     }
//     return (
//         <div className={"fullPage"}>
//             <div>
//                 {/* TODO: DONT SHOW BY DEFAULT */"DONT SHOW BY DEFAULT, ONLY FOR ADMINS"}
//                 <TextInput name="adminName" wrapperName="adminUserPassSubArea" labelText="Admin Username" labelClass="loginLabel" inputType="text" placeholderText="Enter Admin Username" inputClass="loginTextBox" value={adminUser || ""} updateTextHandler={setAdminUser} />
//                 <TextInput name="adminPsw" wrapperName="loginUserPassSubArea" labelText="Admin Password" labelClass="loginLabel" inputType="password" placeholderText="Enter Admin Password" inputClass="loginTextBox" value={adminPass || ""} updateTextHandler={setAdminPass} />
//             </div>
//             <div>
//                 <TextInput name="email" wrapperName="userPassSubArea" labelText="Email" labelClass="loginLabel" inputType="text" placeholderText="Enter Email Address" inputClass="loginTextBox" value={email || ""} updateTextHandler={setEmail} />
//                 <TextInput name="uname" wrapperName="userPassSubArea" labelText="Username" labelClass="loginLabel" inputType="text" placeholderText="Enter Admin Username" inputClass="loginTextBox" value={user || ""} updateTextHandler={setUser} />
//                 <TextInput name="psw" wrapperName="userPassSubArea" labelText="Password" labelClass="loginLabel" inputType="password" placeholderText="Enter Password" inputClass="loginTextBox" value={pass || ""} updateTextHandler={setPass} />
//             </div>
//             <div>
//                 GOOGLE LOGIN AREA
//             </div>
//             {/*/!* TODO: ERROR FOR FAILED LOGIN *!/*/}
//             {/*<div className={"loginRow"}>*/}
//             {/*    <ErrorDisplay err={err} headerLevel={1}/>/!* TODO: headerLevel ok? *!/*/}
//             {/*</div>*/}
//             {/*<div className="loginRow loginUserPass">*/}
//             {/*    /!* TODO: STYLE loginRow, loginUserPass, loginUserPassSubArea, loginLabel, loginTextBox*!/*/}
//             {/*    <TextInput name="uname" wrapperName="loginUserPassSubArea" labelText="Username" labelClass="loginLabel" inputType="text" placeholderText="Enter Username" inputClass="loginTextBox" value={user} updateTextHandler={setUser} />*/}
//             {/*    <TextInput name="psw" wrapperName="loginUserPassSubArea" labelText="Password" labelClass="loginLabel" inputType="password" placeholderText="Enter Password" inputClass="loginTextBox" value={pass} updateTextHandler={setPass} />*/}
//
//             {/*    <div className="loginUserPassSubArea">/!* TODO: STYLE*!/*/}
//             {/*        <div className="loginUserPassSubAreaLeft">/!* TODO: STYLE*!/*/}
//             {/*            <label htmlFor="rememberMe" className="loginLabel">/!* TODO: style loginLabel *!/*/}
//             {/*                /!* TODO: style loginRememberMe *!/*/}
//             {/*                {"Remember me:"}<input type="checkbox" checked={remember} name="rememberMe"*/}
//             {/*                                       className="loginRememberMe" onChange={()=>setRemember(!remember)}/>*/}
//             {/*            </label>*/}
//             {/*        </div>*/}
//             {/*        <div className="loginUserPassSubAreaRight">/!* TODO: STYLE*!/*/}
//             {/*            <button type="submit" id="loginSubmit" onClick={tryLoginUserPass}>Login</button>*/}
//             {/*        </div>*/}
//             {/*    </div>*/}
//
//             {/*</div>*/}
//             {/*<div className={"loginRow"}>*/}
//             {/*    <div className="loginGoogle">/!* TODO: style loginGoogle.... https://www.npmjs.com/package/@react-oauth/google*!/*/}
//             {/*        <GoogleLogin*/}
//             {/*            onSuccess={handleGoogleLoginSuccess}*/}
//             {/*            onError={() => {*/}
//             {/*                setErr("Error logging in via google")*/}
//             {/*                return*/}
//             {/*            }}*/}
//             {/*        />;*/}
//             {/*    </div>*/}
//             {/*</div>*/}
//         </div>
//     )
// }

// function SignUpArea({onSignup}:{onSignup:(sessId:string)=>void}) {
//     const searchParams = useSearchParams();
//     // This function will be called upon a successful login
//     const handleSuccess = (credentialResponse: CredentialResponse) => {
//         // If you are using the authorization code flow, you will receive a code to be exchanged for an access token
//         const authorizationCode = credentialResponse.credential;
//
//         // Send the authorization code to your backend server
//
//         fetch(loginDestination('/login', searchParams), { // TODO: FIX THIS ENDPOINT!!!
//             method: 'POST',
//             headers: { // TODO: no auth headers?
//                 'Content-Type': 'application/json',
//             },
//             body: JSON.stringify({ code: authorizationCode }),
//         })
//             .then(response => response.text())
//             .then(onSignup)
//             .catch(error => {
//                 // Handle errors in communicating with your backend server
//                 console.error('Error exchanging authorization code:', error);
//             });
//     };
//     const guestSignIn = ()=>{
//         fetch(loginDestination('/guestLogin', searchParams),{ // TODO:
//             method: 'POST',
//             headers: { // TODO: no auth headers?
//                 'Content-Type': 'application/json',
//             },
//         })
//             .then(response => response.text())
//             .then(onSignup)
//             .catch(error => {
//                 // Handle errors in communicating with your backend server
//                 console.error('Error signing in as guest:', error);
//             });
//     }
//
//     const handleError = () => {
//         console.error('Google login failed');
//     };
//
//     return (
//         <div>
//             <GoogleLogin
//                 onSuccess={handleSuccess}
//                 onError={handleError}
//                 useOneTap
//             />
//             <button className={"basicButtonSmall"} onClick={guestSignIn}>{"Continue as guest"}</button>
//         </div>
//     );
// }

function loginDestination(basePath: string, searchParams: ReadonlyURLSearchParams): string {
    const dst = searchParams.get('destination')
    if (dst !== null) {
        return basePath +'?destination=' + encodeURIComponent(dst)
    }
    return basePath
}

function SignInArea({onLogin}:{onLogin:(sessId:string)=>void}) {
    const searchParams = useSearchParams();
    const [testEmail, setTestEmail] = useState<string>("")
    const dstParam = searchParams.get('destination')
    const loginParams = (dstParam !== null && dstParam !== "" ? '?destination=' + encodeURIComponent(dstParam):"")
    // This function will be called upon a successful login
    const handleSuccess = (credentialResponse: CredentialResponse) => {
        // // If you are using the authorization code flow, you will receive a code to be exchanged for an access token
        // const authorizationCode = credentialResponse.credential; // TODO: ????
        location.assign('/auth/google'+loginParams)
    };
    const guestSignIn = ()=>{
        fetch('/guestLogin'+loginParams, { // TODO: fetch('/guestLogin', {
            method: 'POST',
            headers: { // TODO: no auth headers?
                'Content-Type': 'application/json',
            },
        })
            .then(response => {
                if (response.redirected) {
                    // response.url contains the final destination URL
                    window.location.href = response.url;
                    return;
                }
            })
            .catch(error => {
                // Handle errors in communicating with your backend server
                console.error('Error signing in as guest:', error);
            });
    }
const testSignIn = ()=>{
        if (testEmail ===""){
            console.error("no test email provided!")
            return
        }
        fetch('/testLogin/'+encodeURI(testEmail)+loginParams, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
        })
            .then(response =>{
                if (response.redirected) {
                    // response.url contains the final destination URL
                    window.location.href = response.url;
                    return;
                }
            })
            .catch(error => {
                // Handle errors in communicating with backend server
                console.error('Error signing in as test user:', error);
            });
    }

    const handleError = () => {
        console.error('Login failed!');
    };

    return (
        <div>
            <GoogleLogin
                // allowed_parent_origin={["mush.appli.ng","localhost"]}
                onSuccess={handleSuccess}
                onError={handleError}
                useOneTap
            />
            <button className={"basicButtonSmall"} onClick={guestSignIn}>{"Continue as guest"}</button>
            {/*<div> TODO: reenable for testing only!*/}
            {/*    <div>{"Test user signin area"}</div>*/}
            {/*    <div className={"inlineChildren"}>*/}
            {/*        <SelectorFor options={[*/}
            {/*            "",*/}
            {/*            "testProjAdminA@appli.ng",*/}
            {/*            "testProjAdminB@appli.ng",*/}
            {/*            "testProjAdminC@appli.ng",*/}
            {/*            "testProjWriteA@appli.ng",*/}
            {/*            "testProjWriteB@appli.ng",*/}
            {/*            "testProjWriteC@appli.ng",*/}
            {/*            "testProjReadA@appli.ng",*/}
            {/*            "testProjReadB@appli.ng",*/}
            {/*            "testProjReadC@appli.ng",*/}
            {/*            "testProjNoneA@appli.ng",*/}
            {/*            "testProjNoneB@appli.ng",*/}
            {/*            "testProjNoneC@appli.ng",*/}
            {/*        ]} initial={""} updateParent={setTestEmail} disabled={false}/>*/}
            {/*        <button className={"basicButtonSmall"} onClick={e=>{*/}
            {/*            e.stopPropagation()*/}
            {/*            testSignIn()*/}
            {/*        }}>{"Sign in with test user"}</button>*/}
            {/*    </div>*/}
            {/*</div>*/}
        </div>
    );
}



interface TextInputProps {
    name: string,
    wrapperName: string,
    labelText: string,
    labelClass: string,
    inputType: string,
    placeholderText: string,
    inputClass: string,
    value: string,
    updateTextHandler: (s: string) => void,
}

function TextInput(props: TextInputProps) { // TODO: delete if unused
    return (
        <div className={props.wrapperName}>
            <label htmlFor={props.name} className={props.labelClass}><b>{props.labelText}</b></label>
            <input type={props.inputType} placeholder={props.placeholderText} name={props.name} className={props.inputClass} value={props.value} onChange={(e) => {props.updateTextHandler(e.currentTarget.value)}} required/>
        </div>
    )
}