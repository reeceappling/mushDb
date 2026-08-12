'use client'

import React, {JSX, useContext, useEffect, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {
    AddCreatedQuadColFunction,
    AllEntries,
    OnViewCreatorQuadCol,
    SplitAllEntries
} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import {
    InitialPicsEntries,
    IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
    PicWithNotesForm,
} from "@/app/components/formSubcomponents/picWithNotes";
import {NewTransferArea, TransfersOutDisplay} from "@/app/components/transferClient";
import {
    CreatedLinkFor,
    DisplayFormWrapper,
    DoCreateRequest,
    DoCreateRequestMultipart,
    DoUpdateMultipartRequest,
    ExistingRecentSelector,
    FlexedArea,
    FlexedSinglesGroup,
    ImportDisplayInput,
    ImportEntryFormWrapper,
    IsString,
    ListPageItems,
    ListPageTable,
    ListTableColumn,
    DoMultipartImportRequest,
    NewColumn,
    NewEntryFormWrapper,
    NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey,
    RequiredKey,
    resolvePicsFormData, Subform, setFormFull, viewUrlFor, PopupApp, PopupInfo, DefaultPopupInfo,
} from "@/app/components/common";
import {
    ErrorDisplay,
    GensFormDisplay,
    MostRecentImageDisplay,
    NameArea,
    ParentDisplay,
    PicsDisplay,
} from "@/app/components/formSubcomponents/commonClient";
import {FruitData} from "@/app/components/fruitServer";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {EntryLinkIdWrapper, EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {NewSporePrintForm} from "@/app/components/sporePrintClient";
import {SpeciesData} from "@/app/components/speciesServer";
import ReaderWriterSelector, {ReadRFIDButton, WriteRfidOvcArea} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {
    ExistingSpeciesSubspeciesSelector,
    SpeciesSubspeciesArea
} from "@/app/components/speciesClient";
import TestAndValidate from "@/app/components/testing/untested";
import {AclDisplay, MarshalAcl, TogglableAreaWithDepth, UnmarshalAcl} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {ChildSwabArea, NewSporeSwabForm, SporeSwabSelectorTable} from "@/app/components/sporeSwabClient";
import {SporeSwabData} from "@/app/components/sporeSwabServer";
import {SporePrintData} from "@/app/components/sporePrintServer";
import {OnViewCreatorsQuadColArea} from "@/app/components/formSubcomponents/ovc";
import {CreatedUpdatedDisposedArea} from "@/app/components/commonServer";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {TransferData} from "@/app/components/transferServer";
import {SelectorFor} from "@/app/components/selector";
import {MssData} from "@/app/components/mssServer";
import {MssSelectorTable} from "@/app/components/mssClient";
import {Actions, ActionTypes, SetModalInfoAction, useModalContext} from "@/app/components/formSubcomponents/modalContext/modal";

export function AssertFruit(input: any): asserts input is FruitData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }

    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['species', 'string'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Bag assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // optional simple keys
    const optionalSimpleKeys = new Map<string, string>([
        ['subspecies', 'string'],
        ['genSpore', 'number'],
        ['parentType', 'string'],
        ['parent', 'string'],
        ['disposed', 'number'],
    ])
    for (const [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Bag assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex required keys
    const complexRequiredKeys = new Map<string, (v: any) => boolean>([
        //['acl', IsValidAcl],
    ])
    for (const [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('Plate assertion failure: required key ' + key + ' was not valid');
        }
    }

    // complex optional keys
    const complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
    ])
    for (const [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Fruit assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', IsString],
        ['prints', IsString],
        ['pics', IsValidPicWithNotesIncoming],
        ['notes', IsValidNote],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Bag assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    return
}

export default function FruitDisplay(
    {
        id, readonly, data, headerLevel, openSporesInNewTab, allowPrintCreation, isTopLevel // TODO: change to regular display????
    }: {
        id: string;
        readonly: boolean;
        isTopLevel: boolean;
        data: FruitData;
        headerLevel?: number;
        openSporesInNewTab?: boolean;
        allowPrintCreation?: boolean;
    }) {
    const {dispatch} = useModalContext();
    const [initial, setInitial] = useState(data)

    const [pics, setPics] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.pics))
    const [disposed, setDisposed] = useState(initial.disposed)
    const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
    // Helper states
    const [transfersOut, setTransfersOut] = useState(data.transfersOut || [])
    const [sporePrints, setSporePrints] = useState(data.prints) // TODO: use?
    const [err, setErr] = useState<string | undefined>()
    const [acl, setAcl] = useState<ACL>(initial.acl)
    //const [popupInfo, setPopupInfo] = useState<PopupInfo>(DefaultPopupInfo)
    useEffect(() => {
        console.log("ACL set to: "+JSON.stringify(MarshalAcl(acl))) // TODO: del!
    }, [acl])
    const updateInitial = (updated: FruitData) => {
        setInitial(updated)
        setPics(InitialPicsEntries(updated.pics))
        setDisposed(updated.disposed)
        setNotes(InitialNotesState(updated.notes))
        // Helper states
        setTransfersOut(updated.transfersOut || [])
        setSporePrints(updated.prints || [])
        setAcl(updated.acl)
        setErr(undefined)
    }
    // TODO: fix?
    // const sporePrintsArea = () => {
    //     return <div>
    //         <div>{"Spore Prints: "}</div>
    //         {(sporePrints.length === 0) &&
    //             <div>{"None"}</div>}
    //         {sporePrints.map(spid => {
    //             const b58id = spid
    //             return <div key={b58id}>
    //                 <EntryLink props={{
    //                     linkId: b58id,
    //                     entryType: "sporePrint",
    //                     openInNewTab: openSporesInNewTab
    //                 }}>{spid}</EntryLink>
    //             </div>
    //         })}
    //     </div>
    // }
    const cookies = useContext(CookiesContext)
    const fruitSubmit = async () => {
        // disposed, notes, existing pics
        const formData = new FormData()
        const dataObj: any = { // TODO: ensure const instead of let is ok here!
            notes: notes,
            disposed: disposed,
            acl: MarshalAcl(acl),
        }
        try {
            // Pics
            const picsInfo = resolvePicsFormData(pics)
            const newImages = picsInfo.images
            dataObj.images = picsInfo.obj
            // Set data on form
            setFormFull(formData, dataObj, newImages, undefined, undefined)
        } catch (caught: any) {
            const newErr = JSON.stringify(caught)
            setErr(newErr)
            dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                    header: "Update Failed",
                    text: "failed to set form values: "+newErr,
                    isErr: true
                }})
            return
        }
        try {
            const resp = await DoUpdateMultipartRequest("fruit", initial._id, formData, AssertFruit, allCookies(cookies))
                .catch(e=>{
                    const newErr = "failed to make request: "+JSON.stringify(e)
                    setErr(newErr)
                    dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                            header: "Update Failed",
                            text: newErr,
                            isErr: true
                        }})
                    throw newErr
                })
            try {
                const temp = new FruitData(resp)
                updateInitial(temp)
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Update Success",
                        text: "entry updated successfully",
                        isErr: false
                    }})
                return
            } catch (e){
                const newErr = "failed to parse response: " + JSON.stringify(e)
                setErr(newErr)
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                    header: "Update Failed",
                    text: newErr,
                    isErr: true
                }})
                return
            }
        } catch {
            // do nothing, err already thrown
        }

        // DoUpdateMultipartRequest("fruit", initial._id, formData, AssertFruit, allCookies(cookies))
        //     .catch(e=>{
        //         const newErr = JSON.stringify(e)
        //         setErr("failed to make request: " + newErr)
        //         setPopupInfo({
        //             header: "Update Failed",
        //             text: newErr,
        //             isErr: true
        //         })
        //         throw newErr
        //     })
        //     .then(v => {
        //         updateInitial(new FruitData(v))
        //         setPopupInfo({
        //             header: "Update Success",
        //             text: "entry updated successfully",
        //             isErr: false
        //         })
        //     })
        //     .catch(e => {
        //         const newErr = JSON.stringify(e)
        //         setErr("failed to make request or parse response: " + newErr)
        //         setPopupInfo({ // TODO: handle error parsing response
        //             header: "Update Failed",
        //             text: newErr,
        //             isErr: true
        //         })
        //     })
    }
    const ovcs: ()=>OnViewCreatorQuadCol[] = ()=>{
        const disp = initial.disposed !== undefined
            return [...(!disp?[{
            txt: "Clone Fruit", // TODO: ensure works as expected?
            newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
                return <NewTransferArea idFrom={data._id} typeFrom={"fruit"}
                /*validTypesTo={["plate","slant","jar","stasisTube","bag","fruitingChamber" TODO: ensure comprehensive list]}*/
                                        onCreated={(item: TransferData) => {
                    setTransfersOut([...transfersOut, item._id]) // TODO: ok?
                    onCreate([{
                        typeText: "Transfer",
                        node: <CreatedLinkFor linkId={item._id} typ={"transfer"}/>,
                    }], false)
                }}/>
            },
        },
        {
            txt: "Create Spore Swab",
            newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
                return <NewSporeSwabForm fruitIn={data} onCreate={(item: SporeSwabData) => {
                    onCreate([{
                        typeText: "Spore Swab",
                        node: <CreatedLinkFor linkId={item._id} typ={"sporeSwab"}/>,
                    }], false)
                }}/>
            },
        },
        {
            txt: "Create Spore Print",
            newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
                return <NewSporePrintForm parentId={data._id} parentTypeIn={"fruit"}
                                          onCreate={(item: SporePrintData) => {
                                              setSporePrints([...(sporePrints || []), item._id])
                                              onCreate([{
                                                  typeText: "Spore Print",
                                                  node: <CreatedLinkFor linkId={item._id} typ={"sporePrint"}/>,
                                              }], false)
                                          }}/>
            },
        },
        WriteRfidOvcArea(initial._id),
        ]:[]),

    ]}
    return (
        <DisplayFormWrapper entryType={"fruit"}>
            <ErrorDisplay err={err}/>
            <ID props={{id:data._id, txt:"Fruit", entryType:"fruit", linkPage:false, allowOpenMainPage:false}}/>
            <OnViewCreatorsQuadColArea OnViewCreators={ovcs()} readonly={readonly}/>
            <MostRecentImageDisplay data={initial.mostRecentImage}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                    <ParentDisplay parent={initial.parent} parentType={initial.parentType}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated}
                                                readonly={readonly} initialDisposed={initial.disposed}
                                                setDisposedOnParent={setDisposed}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <GensFormDisplay gensSinceSpore={initial.genSpore} dontDisplayGensFruitOrSpore={true}/>
                </FlexedSinglesGroup>
            </FlexedArea>
            <TransfersOutDisplay thisId={initial._id} thisEntryType={"fruit"} transfersOut={transfersOut}
                                 allowNewTransferCreation={false}/>{/* TODO: validTypesTo*/}
            <FruitPrintsDisplay prints={sporePrints}/>
            <ChildSwabArea parent={initial._id}/>{/* TODO: SWABS DISPLAY?*/}
            <PicsDisplay pix={initial.pics || []} updateParent={setPics} readonly={readonly}/>{/* Pics */}
            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay initial={initial.acl} readonly={readonly} updateParent={setAcl}/>
            </TogglableAreaWithDepth>
            {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                e.stopPropagation();
                fruitSubmit()
            }}>{"Update"}</button>}
        </DisplayFormWrapper>
    )
}

function FruitPrintsDisplay({prints}:{prints?:string[]}){
    if (!prints || prints.length === 0){
        return null
    }
    return <Subform>
        <div className={"text-lg areaHeader"}>{"Spore Prints:"}</div>{/* TODO: text-lg ok?*/}
        <div className={"flex flex-row flex-wrap justify-around items-center gap-2"}>
            {prints.map(id=><EntryLinkIdWrapper key={id} props={{
                linkId: id,
                entryType: "sporePrint",
                openInNewTab: false,
            }}>
                <div className={"fruitPrint p-1"}>{id}</div>
            </EntryLinkIdWrapper>)}
        </div>
    </Subform>
}

export function NewFruitForm(
    {parentId, parentType, headerLevel, readonly, onCreate}: {
        parentId: string,
        parentType: string,
        headerLevel?: number,
        readonly: boolean,
        onCreate: (f: FruitData) => void
    }) {
    if (readonly) {
        return null
    }
    const {dispatch} = useModalContext();
    const [harvestDate, setHarvestDate] = useState(Date.now())
    const [pics, setPics] = useState<NewPicWithNotesForm[]>([])
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>()

    const cookies = useContext(CookiesContext)
    const newFruitSubmit = () => {
        const formData = new FormData()
        const dataObj: any = {
            parentId: parentId,
            parentType: parentType,
            harvestDate: harvestDate,
            notes: notes,
            writeTagTo: writeTagTo,
        }
        if (pics.length > 0) {
            dataObj.pics = pics.map(p => {
                return {
                    time: p.time, notes: p.notes.new.map(v => {
                        return v.data
                    })
                }
            })
            formData.set("data", JSON.stringify(dataObj))
            for (let i = 0; i < pics.length; i++) {
                const imgi = pics[i].img
                if (imgi === undefined) {
                    setErr("new image #" + i + " was not set!")
                    return
                }
                const filePrefix = "newPic" + "-" + i
                formData.set(filePrefix, imgi, filePrefix)
            }
        }
        DoCreateRequestMultipart("fruit", formData, AssertFruit, allCookies(cookies))
            .then((v)=>{
                onCreate(v)
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Creation succeeded",
                        text: "created",
                        isErr: false
                    }})
            })
            .catch(e => {
                const newErr = JSON.stringify(e)
                setErr(newErr)
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Creation failed",
                        text: newErr,
                        isErr: true
                    }})
            })
    }
    return (
        <NewEntryFormWrapper entryType={"fruit"}>
            <ErrorDisplay err={err}/>
            {/* TODO: say harvest date is today?*/}
            <PicsDisplay pix={[]} updateParent={v => {
                setPics(v.new)
            }} headerLevel={headerLevel} readonly={false}/>
            <NewEntryNotes setNotes={setNotes}/>
            <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
            <button className={"greenButton"} onClick={e=>{
                e.stopPropagation()
                newFruitSubmit()
            }}>{"Create Fruit"}</button>
        </NewEntryFormWrapper>
    )
}

export function FruitImportDisplay({headerLevel}: ImportDisplayInput) { // USE ONLY FOR FRUITS PURCHASED OR FOUND
    const {dispatch} = useModalContext();
    const [parentType, setParentType] = useState<string | undefined>(undefined) // TODO: ensure this is everywhere in ts and go. Also set parent type where needed
    const [species, setSpecies] = useState<SpeciesData | undefined>(undefined)
    const [subspecies, setSubspecies] = useState<string | undefined>(undefined)
    const [imageFile, setImageFile] = useState<File | undefined>(undefined)
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const cookies = useContext(CookiesContext)
    const submitImportFruit = () => {
        if (parentType === undefined) {
            setErr("source area must be set!")
            return
        }
        if (parentType !== "store" && parentType !== "outside" && parentType !== "online") {
            setErr("parentType must be store or outside!")
            return
        }
        if (species === undefined) {
            setErr("Species must be set!")
            return
        }
        const formData = new FormData()
        const dataObj: any = {
            parentType: parentType, // "store" or "outside" or "online" TODO: or specify if online?
            species: species._id,
            notes: notes,
            // optional
            subspecies: subspecies,
            writeTagTo: writeTagTo,
        }
        formData.set("data", JSON.stringify(dataObj))
        imageFile && formData.set("img", imageFile, "img")
        const dispatchUpdate = (isErr:boolean, text:string)=>{
            if(isErr){
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Creation failed",
                        text: text,
                        isErr: true
                    }})
            } else {
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Creation successful",
                        text: text,
                        isErr: false
                    }})
            }
        }
        DoMultipartImportRequest(formData, "fruit", AssertFruit, setErr, allCookies(cookies), dispatchUpdate)
    }
    return <ImportEntryFormWrapper entryType={"fruit"}>
        <ErrorDisplay err={err}/>
        {/* Required Fields */}
        <div className={"inlineChildren"}>
            <div>{"Source: "}</div>
            <SelectorFor options={["", "store", "outside"]} initial={""} updateParent={setParentType} disabled={false}/>
        </div>
        {/* TODO: ParentType: FOR "store" OR "outside" ONLY!!!!! */}{/* TODO: THIS!*/}
        <ExistingSpeciesSubspeciesSelector doSelectSpecies={setSpecies} doSelectSubspecies={setSubspecies}/>
        {/*<ExistingSpeciesSelector doSelect={setSpecies} headerLevel={headerLevel}/>*/}
        {/*/!* Optional fields*!/*/}
        {/*<ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies}*/}
        {/*                            headerLevel={headerLevel}/>*/}
        <ImageSelector updateParent={setImageFile}/>
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        {/* SUBMIT AREA */}
        <button className={"bottomButton greenButton"} onClick={submitImportFruit}>{"Import"}</button>
    </ImportEntryFormWrapper>
}

// export function CreateCloneArea( // TODO: this vs NewFruitForm
//     {
//         fruitId, headerLevel, onCloneCreated, readonly,
//     }: {
//         fruitId: string,
//         headerLevel?: number,
//         onCloneCreated: (f: FruitData) => void,
//         readonly: boolean,
//     }) {
//     if (readonly) {
//         return null
//     }
//     const [typeTo, setTypeTo] = useState("plate")
//     const [idTo, setIdTo] = useState<string | undefined>()
//     const [notes, setNotes] = useState<Note[]>([])
//     const [err, setErr] = useState<string | undefined>()
//
//     const cookies = useContext(CookiesContext)
//     const handleCreate = () => {
//         const body: any = {
//             idFrom: fruitId,
//             typeFrom: "fruit",
//             typeTo: typeTo,
//             idTo: idTo,
//             notes: notes,
//             // TODO: writeTagTo?
//         }
//         DoCreateRequest("clone", body, AssertFruit, allCookies(cookies))
//             .then(c => {
//                 onCloneCreated(new FruitData(c))
//             })
//             .catch(e => {
//                 setErr("failed to create/get new clone: " + JSON.stringify(e))
//             })
//     }
//     return <div>
//         <ErrorDisplay err={err} />
//         <div>
//             <div>{"Create Clone:"}</div>
//             <div>
//                 <TestAndValidate todos={["no need for type?"]}>
//                     <div>{"TYPE TO:"}</div>
//                 </TestAndValidate>
//                 <select className={"tailwindSelector"} value={typeTo} onSelect={e => {
//                     setTypeTo(e.currentTarget.value)
//                 }} onChange={() => {
//                 }}>
//                     {["plate", "jar", "slant"].map((opt, i) => {
//                         return <option value={opt} key={i}>{opt}</option>
//                     })}
//                 </select>
//             </div>
//             <div>
//                 <TestAndValidate
//                     todos={["validate that this is working properly in typing as well as reading from rfid"]}>
//                     <NameArea currentName={idTo} setName={setIdTo} headerTxt={"Select ID: "} readonly={false}
//                               headerLevel={headerLevel}/>
//                 </TestAndValidate>
//                 <ReadRFIDButton handleTagRead={setIdTo}/>
//             </div>
//         </div>
//         <NewEntryNotes setNotes={setNotes}/>
//         <button className={"basicButton"} onClick={e => {
//             e.preventDefault()
//             handleCreate()
//         }}>{"Submit new Clone"}</button>
//     </div>
// }

export function FruitListPageTable({data, onClick, withLink}: ListPageItems<FruitData>) {
    let cols: ListTableColumn<FruitData>[] = [
        NewColumn("ID", (v) => v._id, true),
        NewColumn("Harvest", (v) => {
            return NumberToDateStr(v.creationDate)
        }, true),
        NewColumn("Species", v => v.species, true),
        NewColumn("Subspecies", (v) => v.subspecies || "", true),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: FruitData) => {
            return <EntryLinkWrapper props={{entry: v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v => {
        return new FruitData(v)
    }}/>
}

export function FruitSelectorTable({data, onClick}: ListPageItems<FruitData>) {
    return <FruitListPageTable data={data} onClick={onClick} withLink={true}/>
}

export function FruitSelector(
    {
        doSelect,
        hideDisposed = false
    }: {
        doSelect: (val: FruitData | undefined) => void,
        hideDisposed?: boolean
    }) {
    const table = (items: FruitData[]): JSX.Element => {
        return <FruitSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"fruit"} entryTypes={"fruits"} doSelect={doSelect} asserter={AssertFruit}
                                   table={table} hideDisposed={hideDisposed}/>
}