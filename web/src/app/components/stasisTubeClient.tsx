'use client'

import React, {JSX, useContext, useState} from "react";
import {
    IsValidNote,
    NewEntryNotes,
    Note,
    NotesFormArea
} from "@/app/components/formSubcomponents/notes";
import {
    AllEntries,
    Data,
    OnViewCreatorQuadCol,
    SplitAllEntries
} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {
    InitialPicsEntries, IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
    PicWithNotesForm
} from "@/app/components/formSubcomponents/picWithNotes";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {InnocDisplay, TransfersOutDisplay} from "@/app/components/transferClient";
import {KnownFruitableArea} from "@/app/components/formSubcomponents/knownFruitableArea";
import {GenerationInput} from "@/app/components/formSubcomponents/generationInput";
import {
    DisplayFormWrapper,
    DisplayInput,
    DoCreateRequest,
    DoUpdateRequest,
    ExistingRecentSelector,
    FlexedArea,
    FlexedSinglesGroup,
    ImportEntryFormWrapper,
    ListPageItems,
    ListPageTable,
    ListTableColumn,
    DoMultipartImportRequest,
    NewColumn,
    NewEntryFormWrapper,
    NewEntryInput,
    NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey,
    RequiredKey,
    resolveContamsFormData,
    resolvePicsFormData,
    setFormFull,
} from "@/app/components/common";
import ReaderWriterSelector, {WriteRfidOvcArea} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {
    ErrorDisplay,
    GensFormDisplay, MostRecentImageDisplay,
    ParentDisplay,
    PicsDisplay,
} from "@/app/components/formSubcomponents/commonClient";
import {
    ContaminationForm, ContamsDisplay, InitialContamState, IsValidContamination,
    NewContaminationForm
} from "@/app/components/formSubcomponents/contaminations";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {StasisTubeData} from "@/app/components/stasisTubeServer";
import {PcRunArea} from "@/app/components/pcRunClient";
import {PcRunData, PcRunSelectorCloseable} from "@/app/components/pcRunServer";
import {SpeciesData} from "@/app/components/speciesServer";
import {SaleArea} from "@/app/components/saleClient";
import {
    ExistingSpeciesSubspeciesSelector,
    SpeciesSubspeciesArea
} from "@/app/components/speciesClient";
import {AclDisplay, MarshalAcl, TogglableAreaWithDepth, UnmarshalAcl} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {OnViewCreatorsQuadColArea} from "@/app/components/formSubcomponents/ovc";
import {CreatedUpdatedDisposedArea} from "@/app/components/commonServer";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {ActionTypes, useModalContext} from "@/app/components/formSubcomponents/modalContext/modal";

export function AssertStasisTube(input: any): asserts input is StasisTubeData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('StasisTube assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // optional simple keys
    const optionalSimpleKeys = new Map<string, string>([
        ['pcRun', 'string'],
        ['waterSource', 'string'],
        ['species', 'string'],
        ['subspecies', 'string'],
        ['innoc', 'string'],
        ['genSpore', 'number'],
        ['genFruitOrSpore', 'number'],
        ['parentType', 'string'],
        ['parent', 'string'],
        ['knownFruitable', 'boolean'],
        ['sale', 'string'],
        ['disposed', 'number'],
    ])
    for (const [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('StasisTube assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex required keys
    const complexRequiredKeys = new Map<string, (v: any) => boolean>([
        //['acl', IsValidAcl]
    ])
    for (const [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('Stasis Tube assertion failure: required key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    const complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
    ])
    for (const [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('StasisTube assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', (item) => {
            return typeof item === 'string'
        }],
        ['pics', IsValidPicWithNotesIncoming],
        ['contamination', IsValidContamination],
        ['notes', IsValidNote],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('StasisTube assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    return
}

export default function StasisTubeDisplay(
    {
        readonly, data, headerLevel, isTopLevel
    }: DisplayInput<StasisTubeData>) {
    const {dispatch} = useModalContext();
    const [initial, setInitial] = useState(data)
        const existingNotes: Note[] = initial.notes || []
        const initNotes: Data<Note>[] = existingNotes.map((n) => {
            return {data: n, disabled: false}
        })

        const [images, setImages] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.pics))
        const [contams, setContams] = useState<SplitAllEntries<ContaminationForm, NewContaminationForm>>(InitialContamState(initial.contamination))
        const [knownFruitable, setKnownFruitable] = useState(initial.knownFruitable)
        const [disposed, setDisposed] = useState(initial.disposed)
        const [sale, setSale] = useState(initial.sale)
        const [notes, setNotes] = useState<AllEntries<Note>>({existing:initNotes,new:[]})
        // State helpers
        const [transfersOut, setTransfersOut] = useState<string[]>(initial.transfersOut || [])
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL>(initial.acl)
        const updateInitial = (updated: StasisTubeData)=>{
            setInitial(updated)
            setImages(InitialPicsEntries(updated.pics))
            setContams(InitialContamState(updated.contamination))
            setKnownFruitable(updated.knownFruitable)
            setSale(updated.sale)
            setDisposed(updated.disposed)
            setNotes(InitialNotesState(updated.notes))
            // Helper states
            setTransfersOut(updated.transfersOut || [])
            setAcl(updated.acl)
            setErr(undefined)
        }
        const cookies = useContext(CookiesContext)
        const stasisTubeSubmit = () => {
            const formData = new FormData()
            const dataObj:any={
                knownFruitable: knownFruitable,
                sale: sale,
                disposed: disposed,
                notes: notes,
                acl: MarshalAcl(acl),
            }
            try {
                // Pics
                const picsInfo = resolvePicsFormData(images)
                const newImages = picsInfo.images
                dataObj.images = picsInfo.obj
                // Contams
                const contamsInfo = resolveContamsFormData(contams)
                const newContams = contamsInfo.images
                dataObj.contams = contamsInfo.obj
                // Set data on form
                setFormFull(formData, dataObj, newImages, newContams, undefined)
                // formData.set("data", JSON.stringify(dataObj))
                // setFormImages("newPic", formData, newImages)
                // setFormImages("newContam", formData, newContams)
            } catch (caught: any) {
                setErr(JSON.stringify(caught))
                return
            }

            DoUpdateRequest("stasisTube",data._id, formData, AssertStasisTube, allCookies(cookies))
                .then(v=>{
                    updateInitial(new StasisTubeData(v))
                    dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                            header: "Update Success",
                            text: "entry updated successfully",
                            isErr: false
                        }})
                })
                .catch(e=>{
                    setErr("failed to update initial: "+JSON.stringify(e))
                    dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                            header: "Update Failed",
                            text: "failed to update: " + JSON.stringify(e),
                            isErr: true
                        }})
                })
        }
    const ovcs: ()=>OnViewCreatorQuadCol[] = ()=> {
        const disp = initial.disposed !== undefined
        return !disp ? [
            WriteRfidOvcArea(initial._id),
        ]:[]
    }
    const isInnoculated = ()=>{
        return initial.species !== undefined
    }
        return (
            <DisplayFormWrapper entryType={"stasisTube"}>
                <ErrorDisplay err={err}/>
                <ID props={{id:data._id, txt:"Stasis Tube", entryType:"stasisTube"}}/>
                <OnViewCreatorsQuadColArea OnViewCreators={ovcs()} readonly={readonly}/>
                <MostRecentImageDisplay data={initial.mostRecentImage}/>
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated} initialDisposed={initial.disposed} setDisposedOnParent={setDisposed} readonly={readonly}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <PcRunArea binaryId={initial.pcRun}/>
                        {isInnoculated()&&<KnownFruitableArea initial={knownFruitable} doSelect={setKnownFruitable} readonly={readonly} headerLevel={headerLevel}/>}
                        {isInnoculated()&&<SaleArea sale={sale} setSale={setSale} readonly={readonly} canCreateSale={true}/>}
                    </FlexedSinglesGroup>
                    {isInnoculated()&&<FlexedSinglesGroup>
                        <GensFormDisplay gensSinceSpore={initial.genSpore} gensSinceFruitOrSpore={initial.genFruitOrSpore}/>
                    </FlexedSinglesGroup>}
                    {isInnoculated()&&<FlexedSinglesGroup>
                        <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>{/* TODO: allow changing subspecies for mainCollectionItems at some point???*/}
                        <InnocDisplay innoc={initial.innoc} openInNewTab={false}/>
                        <ParentDisplay parent={initial.parent} parentType={initial.parentType}/>
                    </FlexedSinglesGroup>}
                </FlexedArea>
                {isInnoculated()&&<TransfersOutDisplay thisId={initial._id} thisEntryType={"stasisTube"} allowNewTransferCreation={!readonly} transfersOut={transfersOut}
                                     disposeAfter={true} headerTxt={"Transfers"}/>}
                <PicsDisplay pix={initial.pics || []} updateParent={setImages} readonly={readonly} headerLevel={headerLevel} />{/* Pics */}
                <ContamsDisplay initial={initial.contamination || []} updateParent={setContams} readonly={readonly} headerLevel={headerLevel}/>
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                    <AclDisplay initial={initial.acl} readonly={readonly} updateParent={setAcl} />
                </TogglableAreaWithDepth>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
                    e.stopPropagation();
                    stasisTubeSubmit()
                }}>{"Update"}</button>}
            </DisplayFormWrapper>
        )
}

export function NewStasisTubeForm({handlers, pcRunIn}: {handlers: NewEntryInput<StasisTubeData>, pcRunIn?: PcRunData}){
    const {dispatch} = useModalContext();
    const [pcRun, setPcRun] = useState<PcRunData | undefined>(pcRunIn)
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)

    const cookies = useContext(CookiesContext)
    const createStasisTube = (e: React.MouseEvent)=>{
        e.preventDefault()
        if(pcRun===undefined){
            setErr("pc run must be defined")
            return
        }
        const body:any={
            // TODO; consider adding optional water jar field
            pcRun: pcRun._id,
            notes: notes,
            writeTagTo:writeTagTo,
        }
        DoCreateRequest("stasisTube", body, AssertStasisTube, allCookies(cookies))
            .then(v=>{
                handlers.onCreate ? handlers.onCreate(new StasisTubeData(v)) : console.log("no onCreate provided")
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Create Success",
                        text: "entry created successfully",
                        isErr: false
                    }})
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Create Failure",
                        text: "entry failed to create: " + JSON.stringify(e),
                        isErr: true
                    }})
            })
    }
    return <NewEntryFormWrapper entryType={"stasisTube"}>
        <ErrorDisplay err={err} />
        {pcRunIn !== undefined && <PcRunSelectorCloseable doSelect={setPcRun} allowCreation={handlers.isTopLevel} creatorInPage={true}/>}
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo} />
        <button className={"greenButton"} onClick={createStasisTube}>{"Create"}</button>
    </NewEntryFormWrapper>
}

export function StasisTubeImportDisplay() {
    const {dispatch} = useModalContext();
    const [created, setCreated] = useState<number>(Date.now())
    const [species, setSpecies] = useState<SpeciesData | undefined>(undefined)
    const [subspecies, setSubspecies] = useState<string | undefined>(undefined)
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>(undefined)
    const [generation, setGeneration] = useState<number | undefined>(1)
    const [imageFile, setImageFile] = useState<File | undefined>(undefined)
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const cookies = useContext(CookiesContext)
    const importEntry = () => {
        const formData = new FormData() // TODO: const ok?
        const dataObj: any = {
            creationDate:created,
            // optional
            species: species?._id,
            subspecies: subspecies,
            knownFruitable: knownFruitable,
            generation: generation,
            notes: notes,
            writeTagTo: writeTagTo,
        }
        formData.set("data", JSON.stringify(dataObj))
        if(imageFile!==undefined){
            formData.set("image", imageFile, "img")
        }
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
        DoMultipartImportRequest(formData, "stasisTube", AssertStasisTube, setErr, allCookies(cookies), dispatchUpdate)
    }
    return <ImportEntryFormWrapper entryType={"stasisTube"}>
        {err!=undefined && <div>{"Error: "+err}</div>}
        <DateArea pre={"Created: "} when={created} readonly={false} updateParent={setCreated}/>
        <ExistingSpeciesSubspeciesSelector doSelectSpecies={setSpecies} doSelectSubspecies={setSubspecies}/>
        <KnownFruitableArea doSelect={setKnownFruitable}/>
        <GenerationInput updateParent={setGeneration}/>
        <ImageSelector updateParent={setImageFile}/>
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo} />
        <button className={"greenButton"} onClick={importEntry}>{"Import Stasis Tube"}</button>
    </ImportEntryFormWrapper>
}

export function StasisTubeListPageTable({data, onClick, withLink}: ListPageItems<StasisTubeData>) {
    let cols: ListTableColumn<StasisTubeData>[] = [
        NewColumn("ID", (v)=>v._id, true),
        NewColumn("Created", (v)=>{
            return NumberToDateStr(v.creationDate)
        }, true),
        NewColumn("Spec", (v)=>v.species||"", true),
        NewColumn("Subspec", v=>v.subspecies||"", true),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }), // TODO: fit?
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: StasisTubeData)=>{
            return <EntryLinkWrapper props={{entry:v,openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new StasisTubeData(v)}}/>
}
export function StasisTubeSelectorTable({data, onClick}: ListPageItems<StasisTubeData>) {
    return <StasisTubeListPageTable data={data} onClick={onClick} withLink={true} />
}
// export function StasisTubeSelector(
//     {
//         doSelect,
//         allowCreate,
//         hideDisposed
//     }: {
//         doSelect: (val: StasisTubeData | undefined) => void,
//         allowCreate?: boolean,
//         hideDisposed?:boolean
//     }) {
//     const table = (items: StasisTubeData[]):JSX.Element=>{
//         return <StasisTubeSelectorTable data={items} onClick={doSelect}/>
//     }
//
//     return <ExistingRecentSelector entryType={"stasisTube"} entryTypes={"stasisTubes"} doSelect={doSelect} asserter={AssertStasisTube}
//                                    table={table} hideDisposed={hideDisposed}>
//         {allowCreate && <NewStasisTubeForm handlers={{onCreate: doSelect,isTopLevel: false}}/>}
//     </ExistingRecentSelector>
// }