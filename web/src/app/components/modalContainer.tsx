import {
    Actions,
    ActionTypes,
} from "@/app/components/formSubcomponents/readerWriterButtons/readerOptsContext";
import {Dispatch} from "react";

// TODO: THIS IS A SERVER COMPONENT!
export default async function ModalContainer({
                                                 modalType,
                                                 recordId,
                                                 dispatch
}:{
    modalType: string,
    recordId: string,
    dispatch: Dispatch<Actions>}
) {
    let linkTo = <div>LINK TO REAL PAGE HERE</div> // TODO: TEMP
    let pg = <div>PAGE HERE</div> // TODO: TEMP
    let subParams = {id: recordId, readonly: true}
    switch (modalType) {
        // case "pcRun":
        //     // TODO: linkTo
        //     pg = <PcRunDisplay params={subParams}/>
        //     break;
        // case "agarBatch":
        //     // TODO: linkTo
        //     pg = <AgarBatchDisplay params={subParams}/>
        //     break;
        // case "agarRecipe":
        //     // TODO: linkTo
        //     pg = <AgarRecipeDisplay params={subParams}/>
        //     break;
        // // TODO: ALL OTHER MAIN DISPLAYS
        // // TODO: lists!
        // // Recents only
        // case "agarBatchList":
        //     // TODO: LinkTo
        //     pg = <div>LIST OF RECENT</div> // TODO: FIXME
        //     break;
        // case "bagList":
        //     // TODO: LinkTo
        //     pg = <div>LIST OF RECENT</div> // TODO: FIXME
        //     break;
        // case "fruitList":
        //     // TODO: LinkTo
        //     pg = <div>LIST OF RECENT</div> // TODO: FIXME
        //     break;
        // case "boxList":
        // case "fruitingChamberList":
        //     // TODO: LinkTo
        //     pg = <div>LIST OF RECENT</div> // TODO: FIXME
        //     break;
        // case "jarList":
        //     // TODO: LinkTo
        //     pg = <div>LIST OF RECENT</div> // TODO: FIXME
        //     break;
        // case "lcList":
        // case "liquidCultureList":
        //     // TODO: LinkTo
        //     pg = <div>LIST OF RECENT</div> // TODO: FIXME
        //     break;
        // case "mssList":
        //     // TODO: LinkTo
        //     pg = <div>LIST OF RECENT</div> // TODO: FIXME
        //     break;
        // case "pcRunList":
        //     // TODO: LinkTo
        //     pg = <div>LIST OF RECENT</div> // TODO: FIXME
        //     break;
        // case "plateList":
        //     // TODO: LinkTo
        //     pg = <div>LIST OF RECENT</div> // TODO: FIXME
        //     break;
        // case "slantList":
        //     // TODO: LinkTo
        //     pg = <div>LIST OF RECENT</div> // TODO: FIXME
        //     break;
        // case "sporePrintList":
        //     // TODO: LinkTo
        //     pg = <div>LIST OF RECENT</div> // TODO: FIXME
        //     break;
        // case "stasisTubeList":
        //     // TODO: LinkTo
        //     pg = <div>LIST OF RECENT</div> // TODO: FIXME
        //     break;
        // case "transferList":
        //     // TODO: LinkTo
        //     pg = <div>LIST OF RECENT</div> // TODO: FIXME
        //     break;
        //
        // // Recents and standards
        // case "agarRecipeList":
        //     // TODO: LinkTo
        //     pg = <div>LIST OF RECENT AND STANDARDS</div> // TODO: FIXME
        //     break;
        // case "jarRecipeList":
        //     // TODO: LinkTo
        //     pg = <div>LIST OF RECENT AND STANDARDS</div> // TODO: FIXME
        //     break;
        // case "substrateRecipeList":
        //     // TODO: LinkTo
        //     pg = <div>LIST OF RECENT AND STANDARDS</div> // TODO: FIXME
        //     break;
        // // Full Lists
        //
        // case "speciesList":
        //     // TODO: LinkTo
        //     pg = <div>LIST OF ALL</div> // TODO: FIXME
        //     break;
        // case "subspeciesList":
        //     // TODO: LinkTo
        //     pg = <div>LIST OF ALL</div> // TODO: FIXME
        //     break;
        default:
            // TODO: linkTo
            pg = <div>UNHANDLED SUB-PAGE TYPE</div> // TODO: ????????
    }


    return (
        <div
            className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full flex items-center justify-center"> {/* TODO: ENSURE STYLING RIGHT */}
            <div className="p-8 border w-96 shadow-lg rounded-md bg-white"> {/* TODO: ENSURE STYLING RIGHT */}
                <div className="text-center"> {/* TODO: ENSURE STYLING RIGHT */}
                    <h3 className="text-2xl font-bold text-gray-900">Modal Title</h3> {/* TODO: ENSURE STYLING RIGHT */}
                    <div className="mt-2 px-7 py-3"> {/* TODO: ENSURE STYLING RIGHT */}
                        <p className="text-lg text-gray-500">{"Modal Body"}</p>
                    </div>
                    <div className="flex justify-center mt-4">

                        {/* Using useRouter to dismiss modal*/}
                        <button
                            onClick={()=>{dispatch({type: ActionTypes.SET_MODAL_INFO})}}
                            className="px-4 py-2 bg-blue-500 text-white text-base font-medium rounded-md shadow-sm hover:bg-gray-400 focus:outline-none focus:ring-2 focus:ring-gray-300"
                        >
                            Close
                        </button>
                        {/* TODO: LOAD PAGE */}
                        {pg}
                    </div>
                </div>
            </div>
        </div>
    );
}