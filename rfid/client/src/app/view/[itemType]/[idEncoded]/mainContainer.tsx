// 'use client'
//
// import ModalController from "@/app/components/modalController";
// import {JSX, useState} from "react";
// import TopBar from "@/app/components/TopBar";
// import {
//     ModalInfo,
//     ReaderOptionsContextProvider
// } from "@/app/components/formSubcomponents/readerWriterButtons/readerOptsContext";
// import SpeciesDisplay from "@/app/components/speciesClient";
// import SubstrateRecipeDisplay from "@/app/components/substrateRecipeClient";
//
// export default function ClientContainer({entryType,entryData,entryId,rfidReaders}: {
//     entryType: string,
//     entryData: any,
//     entryId: string,
//     rfidReaders: string[],
// }) {
//     const [modalInfo, setModalInfo] = useState<ModalInfo | undefined>(undefined) // TODO: MOVE MODAL STUFF TO APP CONTEXT
//     // TODO: MainContainer contains/controls topBar (client), form container (client), modal controller (client), all wrapped in RFID context.
//     // const finalParams = await params
//     // const id = finalParams.id
//     let displayArea = <div>{"LOADING"}</div>
//     switch (entryType) {
//         // case "agarBatch":
//         //     displayArea = <AgarBatchDisplay params={{id: id, readonly: false, setModalInfo: setAndShowModalInfo}}/>
//         //     break;
//         // case "agarRecipe":
//         //     displayArea = <AgarRecipeDisplay params={{id: id, readonly: false, setModalInfo: setAndShowModalInfo}}/>
//         //     break;
//         // case "bag":
//         //     displayArea = <BagDisplay params={{id: id, readonly: false, setModalInfo: setAndShowModalInfo}}/>
//         //     break;
//         // case "fruit":
//         //     displayArea = <FruitDisplay params={{id: id, readonly: false, setModalInfo: setAndShowModalInfo}}/>
//         //     break;
//         // case "fruitingChamber":
//         //     displayArea = <FruitingChamberDisplay params={{id: id, readonly: false, setModalInfo: setAndShowModalInfo}}/>
//         //     break;
//         // case "jar":
//         //     displayArea = <JarDisplay params={{id: id, readonly: false, setModalInfo: setAndShowModalInfo}}/>
//         //     break;
//         // case "jarRecipe":
//         //     displayArea = <JarRecipeDisplay params={{id: id, readonly: false, setModalInfo: setAndShowModalInfo}}/>
//         //     break;
//         // case "lc":
//         //     displayArea = <LcDisplay params={{id: id, readonly: false, setModalInfo: setAndShowModalInfo}}/>
//         //     break;
//         // case "lcRecipe":
//         //     displayArea = <LcRecipeDisplay params={{id: id, readonly: false, setModalInfo: setAndShowModalInfo}}/>
//         //     break;
//         // case "mss":
//         //     displayArea = <MssDisplay params={{id: id, readonly: false, setModalInfo: setAndShowModalInfo}}/>
//         //     break;
//         // case "pcRun":
//         //     displayArea = <PcRunDisplay params={{id: id, readonly: false, setModalInfo: setAndShowModalInfo}}/>
//         //     break;
//         // case "plate":
//         //     displayArea = <PlateDisplay params={{id: id, readonly: false, setModalInfo: setAndShowModalInfo}}/>
//         //     break;
//         // case "slant":
//         //     displayArea = <SlantDisplay params={{id: id, readonly: false, setModalInfo: setAndShowModalInfo}}/>
//         //     break;
//         case "species":
//             displayArea = <SpeciesDisplay id={entryId} readonly={false} data={entryData} isTopLevel={true} headerLevel={1}/>
//             break;
//         // case "sporePrint":
//         //     displayArea = <SporePrintDisplay params={{id: id, readonly: false, setModalInfo: setAndShowModalInfo}}/>
//         //     break;
//         // case "stasisTube":
//         //     displayArea = <StasisTubeDisplay params={{id: id, readonly: false, setModalInfo: setAndShowModalInfo}}/>
//         //     break;
//         // case "subspecies": // TODO: this next!
//         //     displayArea = <SubspeciesDisplay params={{id: id, readonly: false, setModalInfo: setAndShowModalInfo}}/>
//         //     break;
//         case "substrateRecipe":
//             displayArea = <SubstrateRecipeDisplay id={entryId} readonly={false} data={entryData} isTopLevel={true} headerLevel={1}/>
//             break;
//         // case "transfer":
//         //     displayArea = <TransferDisplay params={{id: id, readonly: false, setModalInfo: setAndShowModalInfo}}/>
//         //     break;
//         default:
//             return <h1>{"INVALID ITEM TYPE"}</h1>
//     }
//     let modal: JSX.Element | undefined = undefined // TODO: ensure ok
//     if (modalInfo != undefined) {
//         // @ts-ignore
//         // TODO: CHANGE MODAL TO BE CLIENT THAT MANAGES STATE, WHICH CONTAINS ServerComponent that loads data!
//         //let modalOpts = {subpage: modalInfo.modalType, id: modalInfo.recordId, closeFunc: closeModal, setModalInfo: setModalInfo /* TODO: CHANGE SETMODALINFO*/}
//         modal = <ModalController params={{subpage: modalInfo.modalType, id: modalInfo.recordId, closeFunc: closeModal, setModalInfo: setModalInfo /* TODO: CHANGE SETMODALINFO*/}}/>
//     }
//     return (
//         <div>
//             <ReaderOptionsContextProvider initialState={{options: rfidReaders, selected: undefined}}>
//                 <TopBar/>
//                 {displayArea}
//                 {modal}
//                 {/*TODO: <TestingOutput />*/}
//             </ReaderOptionsContextProvider>
//         </div>
//     )
// }