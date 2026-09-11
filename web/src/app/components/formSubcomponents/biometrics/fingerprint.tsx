// 'use client'
//
// import {BaseExternalUrl} from "@/app/components/Constants";
// //import axios from 'axios'; // TODO: remove from project
// import {startRegistration} from '@simplewebauthn/browser';
//
// export default function RegisterFingerprint(){ // TODO: validate working
//     const verifyChallenge = (authResult: any, challenge: any): Promise<void> => {
//         const body = {
//             challenge: challenge,
//             credential: authResult,
//         }
//         return fetch(BaseExternalUrl + '/biometrics/fingerprint/verify-challenge', { // TODO: validate url
//             method: 'POST',
//             headers: {
//                 credentials: 'include',
//                 'Access-Control-Allow-Origin': BaseExternalUrl || "*", // TODO: FIXME! maybe "*"?
//                 'Content-type': "application/json", // TODO: ok?
//                 'Accept': "application/json",
//             },
//             body: JSON.stringify(body),
//         }).then(verifyResp => {
//             if (!verifyResp.ok || verifyResp.status !== 200) {
//                 verifyResp.text().then(txt => {
//                     throw "failed to verify challenge. Status Code " + verifyResp.status + ". Body: " + txt
//                 }).catch(err => {
//                     throw "failed to verify challenge. Status Code " + verifyResp.status + ". Failed to get body"
//                 })
//             }
//         }).catch(err => {
//             if (typeof err == "string") {
//                 throw err
//             } else {
//                 throw JSON.stringify(err)
//             }
//         })
//     }
//     const registerPasskey = () => {
//         fetch(BaseExternalUrl + '/biometrics/fingerprint/register-challenge', // TODO: validate url
//             {
//                 method: 'POST',
//                 headers: {
//                     credentials: 'include',
//                     'Access-Control-Allow-Origin': BaseExternalUrl || "*", // TODO: FIXME! maybe "*"?
//                     'Content-type': "application/json", // TODO: ok?
//                     'Accept': "application/json",
//                 }
//             }
//         ).then(res => {
//             if (!res.ok || res.status !== 200) {
//                 res.text().then(txt => {
//                     throw "response not ok, status " + res.status + ": " + txt
//                 }).catch(err => {
//                     throw err
//                 })
//             }
//             res.json().then(data => {
//                 startRegistration(data).then(authResult => {
//                     verifyChallenge(authResult, data.challenge).then(() => {
//                         alert('Fingerprint registered successfully');
//                     }).catch((err: string) => {
//                         console.error(err);
//                         alert(err);
//                     })
//                 })
//             }).catch(err => {
//                 console.error("failed to register fingerprint: ", JSON.stringify(err));
//                 alert('Fingerprint failed to register');
//             });
//         })
//     }
//         return (
//             <div>
//                 <button onClick={e=>{
//                     e.stopPropagation()
//                     registerPasskey()
//                 }}>{"Register Fingerprint"}</button>
//             </div>
//         );
// // const registerPasskey = async () => {
// //     try {
// //         const res = await axios.post('/api/auth/register-challenge');
// //         if (res.data.code === 200) {
// //             const authResult = await startRegistration(res.data.data);
// //             if (authResult) {
// //                 await axios.post('/api/auth/verify-challenge', {
// //                     challenge: res.data.data.challenge,
// //                     credential: authResult,
// //                 });
// //                 alert('Fingerprint registered successfully');
// //             }
// //         } else {
// //             alert(res.data.message);
// //         }
// //     } catch (err) {
// //         console.error(err);
// //     }
// // };
//
// }
// ;