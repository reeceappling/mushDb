'use client'

// import {
//     ActionTypes,
//     ModalInfo,
//     useRfidReaderContext
// } from "@/app/components/formSubcomponents/readerWriterButtons/readerOptsContext";
// import ModalContainer from "@/app/components/modalContainer";
// import {useEffect} from "react";
//
// export default function ModalController({}: {}) {
//     const {state, dispatch} = useRfidReaderContext()
//     if (state.modalInfo == undefined) {
//         return null
//     }
//     // Set mount/unmounter on render?
//     useEffect(() => { // TODO: is this ok????
//         const handleEsc = (event: KeyboardEvent) => {
//             if (event.key === 'Escape') {
//                 dispatch({type: ActionTypes.SET_MODAL_INFO})
//             }
//         };
//         window.addEventListener('keydown', handleEsc);
//
//         return () => {
//             window.removeEventListener('keydown', handleEsc); // TODO: ensure this is ok!!!
//         };
//     }, []);
//     const modalInfo = state.modalInfo as ModalInfo
//     return <ModalContainer  modalType={modalInfo.modalType} recordId={modalInfo.recordId} dispatch={dispatch}/>
// }