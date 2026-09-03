'use client'

import {JarData} from "@/app/components/jarServer";
import {
    DisplayFormWrapper,
    DisplayInput,
    DoCreateRequest,
    DoUpdateMultipartRequest,
    ExistingRecentSelector,
    FlexedArea,
    FlexedSinglesGroup,
    ImportDisplayInput,
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
    SelectorWrapper,
    setFormFull, CreatedLinkFor,
} from "@/app/components/common";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import React, {JSX, useContext, useState} from "react";
import DateArea from "@/app/components/formSubcomponents/date";
import ID from "@/app/components/formSubcomponents/id";
import {
    ErrorDisplay,
    GensFormDisplay,
    MostRecentImageDisplay,
    ParentDisplay,
    PicsDisplay
} from "@/app/components/formSubcomponents/commonClient";
import {JarRecipeArea, JarRecipeSelector} from "@/app/components/jarRecipeClient";
import {KnownFruitableArea} from "@/app/components/formSubcomponents/knownFruitableArea";
import {PcRunArea} from "@/app/components/pcRunClient";
import {JarRecipeData} from "@/app/components/jarRecipeServer";
import {PcRunData, PcRunSelectorCloseable} from "@/app/components/pcRunServer";
import ReaderWriterSelector, {
    WriteRfidOvcArea
} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {InnocDisplay, TransfersOutDisplay} from "@/app/components/transferClient";
import {
    InitialPicsEntries,
    IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
    PicWithNotesForm
} from "@/app/components/formSubcomponents/picWithNotes";
import {
    ContaminationForm,
    ContamsDisplay,
    InitialContamState,
    IsValidContamination,
    NewContaminationForm
} from "@/app/components/formSubcomponents/contaminations";
import {SaleArea} from "@/app/components/saleClient";
import {
    AddCreatedQuadColFunction,
    AllEntries,
    OnViewCreatorQuadCol,
    SplitAllEntries
} from "@/app/components/formSubcomponents/shared";
import {SpeciesData} from "@/app/components/speciesServer";
import {GenerationInput} from "@/app/components/formSubcomponents/generationInput";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {
    ExistingSpeciesSubspeciesSelector,
    SpeciesSubspeciesArea
} from "@/app/components/speciesClient";
import {JarSizeSelector} from "@/app/components/formSubcomponents/utils/volumeSelector";
import {
    AclDisplay,
    MarshalAcl,
    TogglableAreaWithDepth,
    UnmarshalAcl
} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {GrainBatchData, GrainBatchSelectorCloseable} from "@/app/components/grainBatchServer";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {CreatedUpdatedDisposedArea} from "@/app/components/commonServer";
import {OnViewCreatorsQuadColArea} from "@/app/components/formSubcomponents/ovc";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {WetnessDisplay} from "@/app/components/bagClient";
import {SliderOnlyIfUndefinedWithOpenButton} from "@/app/components/formSubcomponents/utils/slider";
import {ActionTypes, useModalContext} from "@/app/components/formSubcomponents/modalContext/modal";
import {NewFruitingChamberForm} from "@/app/components/fruitingChamberClient";
import {NewLcSyringeForm} from "@/app/components/lcSyringeClient";
import {LcSyringeData} from "@/app/components/lcSyringeServer";
import {FruitingChamberData} from "@/app/components/fruitingChamberServer";
import {NewSporePrintForm} from "@/app/components/sporePrintClient";
import {SporePrintData} from "@/app/components/sporePrintServer";
import {NewSporeSwabForm} from "@/app/components/sporeSwabClient";
import {SporeSwabData} from "@/app/components/sporeSwabServer";

export function AssertJar(input: any): asserts input is JarData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        //['recipe', 'string'],
        ['sizeCups', 'number'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
        ['sizeCups', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Jar assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // optional simple keys
    const optionalSimpleKeys = new Map<string, string>([
        ['pcRun', 'string'],
        ['recipe', 'string'], // TODO: Should be required, but some old ones have it missing
        ['burstGrains', 'number'],
        ['wetness', 'number'],
        ['species', 'string'],
        ['subspecies', 'string'],
        ['innoc', 'string'],
        ['genSpore', 'number'],
        ['genFruitOrSpore', 'number'],
        ['parentType', 'string'],
        ['parent', 'string'],
        ['knownFruitable', 'boolean'],
        ['sale', 'string'],
        ['grainBatch', 'string'],
        ['disposed', 'number'],
    ])
    for (const [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Jar assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex required keys
    const complexRequiredKeys = new Map<string, (v: any) => boolean>([
        //['acl', IsValidAcl]
    ])
    for (const [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('Jar assertion failure: required key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    const complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
    ])
    for (const [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Jar assertion failure: optional key ' + key + ' was not valid');
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
            throw new Error('Jar assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    return
}

export function JarImportDisplay({}: ImportDisplayInput) {
    const {dispatch} = useModalContext()
    const [created, setCreated] = useState<number>(Date.now())
    const [recipe, setRecipe] = useState<JarRecipeData | undefined>()
    const [sizeCups, setSizeCups] = useState<number>(4)
    const [species, setSpecies] = useState<SpeciesData | undefined>()
    const [subspecies, setSubspecies] = useState<string | undefined>(undefined)
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>()
    const [generation, setGeneration] = useState<number | undefined>(1)
    const [wetness, setWetness] = useState<number | undefined>()
    const [burstGrains, setBurstGrains] = useState<number | undefined>()
    const [imageFile, setImageFile] = useState<File | undefined>()
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()
    const cookies = useContext(CookiesContext)
    const importEntry = () => {
        const formData = new FormData()
        const dataObj: any = {
            creationDate: created,
            sizeCups: sizeCups,
            recipe: recipe?._id,
            // optional
            species: species?._id,
            subspecies: subspecies,
            knownFruitable: knownFruitable,
            generation: generation,
            wetness: wetness,
            burstGrains: burstGrains,
            notes: notes,
            writeTagTo: writeTagTo,
        }
        formData.set("data", JSON.stringify(dataObj))
        if (imageFile !== undefined) {
            formData.set("img", imageFile, "img")
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
        DoMultipartImportRequest(formData, "jar", AssertJar, setErr, allCookies(cookies), dispatchUpdate)
    }
    return <ImportEntryFormWrapper entryType={"jar"}>
        {err != undefined && <div>{"Error: " + err}</div>}
        <DateArea pre={"Created: "} when={created} readonly={false} updateParent={setCreated}/>
        <div className={"inlineChildren"}>
            <div className={"mr-2"}>{"Size: "}</div>
            <JarSizeSelector onChange={(s: string) => {
                if (s === "pint") {
                    setSizeCups(2)
                } else if (s === "quart") {
                    setSizeCups(4)
                } else {
                    setErr("invalid size cups")
                }
            }}/>
        </div>
        <SelectorWrapper current={recipe} title={"Jar Recipe"} nameFunc={(v: JarRecipeData) => v._id}>
            <JarRecipeSelector doSelect={setRecipe} allowCreate={false}/>
        </SelectorWrapper>
        <ExistingSpeciesSubspeciesSelector doSelectSpecies={setSpecies} doSelectSubspecies={setSubspecies}/>
        {species && <>
            <KnownFruitableArea doSelect={setKnownFruitable}/>
            <GenerationInput updateParent={setGeneration}/>
        </>}
        <SliderOnlyIfUndefinedWithOpenButton text={"(Optional) Wetness"}defaultValue={5} onChange={setWetness}/>
        <SliderOnlyIfUndefinedWithOpenButton text={"(Optional) Burst Grains"} defaultValue={0} onChange={setBurstGrains}/>

        <ImageSelector updateParent={setImageFile}/>
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton"} onClick={importEntry}>{"Import"}</button>
    </ImportEntryFormWrapper>
}

function sizeFromNum(cups: number) {
    switch (cups) {
        case 1:
            return "cup"
        case 2:
            return "pint"
        case 4:
            return "quart"
        default:
            return "unhandled number of cups (" + cups + ")"
    }
}

function cupsPer(unit: string) {
    switch (unit) {
        case "cup":
            return 1
        case "pint":
            return 2
        case "quart":
            return 4
        case "gallon":
            return 16
        default:
            return -1
    }
}

export default function JarDisplay(
    {
        readonly, data, headerLevel, isTopLevel
    }: DisplayInput<JarData>) {
    const {dispatch} = useModalContext()
        const [initial, setInitial] = useState(data)

        const [knownFruitable, setKnownFruitable] = useState(initial.knownFruitable)
        //const [sale, setSale] = useState(initial.sale)
        const [disposed, setDisposed] = useState(initial.disposed)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [pics, setPics] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.pics))
        const [contams, setContams] = useState<SplitAllEntries<ContaminationForm, NewContaminationForm>>(InitialContamState(initial.contamination))
        const [acl, setAcl] = useState<ACL>(initial.acl)
        const [wetness, setWetness] = useState<number | undefined>(initial.wetness)
        const [burstGrains, setBurstGrains] = useState<number | undefined>(initial.burstGrains)

        // Helper states
        const [transfersOut, setTransfersOut] = useState<string[]>(initial.transfersOut || [])
        const [err, setErr] = useState<string | undefined>()
        const updateInitial = (updated: JarData) => {
            setInitial(updated)
            setKnownFruitable(updated.knownFruitable)
            //setSale(updated.sale)
            setDisposed(updated.disposed)
            setNotes(InitialNotesState(updated.notes))
            setPics(InitialPicsEntries(updated.pics))
            setContams(InitialContamState(updated.contamination))
            setAcl(updated.acl)
            setWetness(updated.wetness) // can only be set once
            setBurstGrains(updated.burstGrains) // can only be set once
            setTransfersOut(updated.transfersOut || [])
            setErr(undefined)
        }
        const cookies = useContext(CookiesContext)
        const submit = () => {
            const formData = new FormData()
            const dataObj: any = {
                knownFruitable: knownFruitable,
                disposed: disposed,
                wetness: wetness,
                burstGrains: burstGrains,
                acl: MarshalAcl(acl),
                notes: notes,
            }
            try {
                // Pics
                const picsInfo = resolvePicsFormData(pics)
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

            DoUpdateMultipartRequest("jar",initial._id, formData, AssertJar, allCookies(cookies))
                .then(v=>{
                    updateInitial(new JarData(v))
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
        const jarSizeArea = () => {
            return <div>
                {"Size: " + sizeFromNum(data.sizeCups)}
            </div>
        }
    const isInnoculated = ()=>{
        return initial.species !== undefined
    }
    const ovcs: ()=>OnViewCreatorQuadCol[] = ()=> {
        const disp = initial.disposed !== undefined
        return !disp ? [
            ...(isInnoculated() ? [/*{
                txt: "New Fruiting Chamber", // TODO: FULLY TEST! New on 8/26/26 // TODO: needs to also propagate species information
                newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
                    return <NewFruitingChamberForm handlers={{isTopLevel: false,onCreate:(fc: FruitingChamberData) => { // TODO: should swap to handler={{}} format rather than direct onCreate
                            onCreate([{
                                typeText: "Fruiting Chamber",
                                node: <CreatedLinkFor linkId={fc._id} typ={"fruitingChamber"}/>,
                            }], false)
                        }}} parent={initial._id} substrateBatchIn={undefined}/>
                },
            },*/{
                txt: "New Spore Print", // TODO: FULLY TEST! New on 8/26/26
                newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
                    return <NewSporePrintForm parentId={initial._id} onCreate={(sp: SporePrintData) => { // TODO: should swap to handler={{}} format rather than direct onCreate
                        onCreate([{
                            typeText: "Spore Print",
                            node: <CreatedLinkFor linkId={sp._id} typ={"sporePrint"}/>,
                        }], false)
                }}/>
                },
            },{
                txt: "New Spore Swab", // TODO: FULLY TEST! New on 8/26/26
                newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
                    return <NewSporeSwabForm otherParentIn={initial._id} onCreate={(sp: SporeSwabData) => { // TODO: should swap to handler={{}} format rather than direct onCreate
                        onCreate([{
                            typeText: "Spore Swab",
                            node: <CreatedLinkFor linkId={sp._id} typ={"sporeSwab"}/>,
                        }], false)
                    }}/>
                },
            },/*{
                // TODO: new fruit?????
            }*/] : []),
            WriteRfidOvcArea(initial._id),
        ] : []
    }

        return <DisplayFormWrapper entryType={"jar"}>
            <ErrorDisplay err={err}/>
            <ID props={{id:data._id, txt:"Grain Jar", entryType:"jar"}}/>
            <MostRecentImageDisplay data={initial.mostRecentImage}/>
            <OnViewCreatorsQuadColArea OnViewCreators={ovcs()} readonly={readonly}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated}
                                                initialDisposed={initial.disposed} setDisposedOnParent={setDisposed}
                                                readonly={readonly}/>
                    {jarSizeArea()}
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    {isInnoculated()&&<SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>}
                    <JarRecipeArea headerLevel={headerLevel} recipeId={initial.recipe}/>
                    <PcRunArea binaryId={initial.pcRun}/>

                </FlexedSinglesGroup>
                {isInnoculated()&&<FlexedSinglesGroup>
                    <ParentDisplay parent={initial.parent} parentType={initial.parentType}/>
                    <InnocDisplay innoc={initial.innoc}/>
                    <KnownFruitableArea initial={knownFruitable} doSelect={setKnownFruitable} readonly={readonly}
                                        headerLevel={headerLevel}/>
                    {/*TODO:???<SaleArea sale={sale} setSale={setSale} readonly={readonly} canCreateSale={true}/>*/}
                </FlexedSinglesGroup>}
                {isInnoculated()&&<FlexedSinglesGroup>
                    <GensFormDisplay gensSinceSpore={initial.genSpore} gensSinceFruitOrSpore={initial.genFruitOrSpore}/>
                </FlexedSinglesGroup>}
            </FlexedArea>
            {initial.wetness===undefined?<SliderOnlyIfUndefinedWithOpenButton defaultValue={5} onChange={setWetness}/> : <WetnessDisplay value={wetness} />}
            {initial.burstGrains===undefined?<SliderOnlyIfUndefinedWithOpenButton text={"Burst Grains"} defaultValue={0} onChange={setBurstGrains}/> : <WetnessDisplay text={"Burst Grains"} value={burstGrains} />}

            {isInnoculated()&&<TransfersOutDisplay thisId={initial._id} thisEntryType={"jar"} transfersOut={transfersOut}
                                 allowNewTransferCreation={!readonly}/>}
            <PicsDisplay pix={initial.pics || []} readonly={readonly} updateParent={setPics}/>
            <ContamsDisplay initial={initial.contamination || []} updateParent={setContams}
                            readonly={readonly} headerLevel={headerLevel}/>

            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay initial={initial.acl} readonly={readonly} updateParent={setAcl}/>
            </TogglableAreaWithDepth>
            {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                e.stopPropagation();
                submit()
            }}>{"Update"}</button>}
        </DisplayFormWrapper>
}

// NewJarForm is used from the recipe page. PcRun CAN be created from here?
// TODO: NewJarForm is used from the grain batch page. PcRun CAN be created from here?
// TODO: Can also be called from the jar recipe page, which will create a new batch as well
export function NewJarForm({handlers, pcRunIn, grainBatchIn}: {
    handlers: NewEntryInput<JarData>,
    pcRunIn?: PcRunData,
    grainBatchIn?: GrainBatchData
}) {
    const {dispatch} = useModalContext()
    //const [creationDate, setCreationDate] = useState(Date.now()) // set serverside
    const [grainBatch, setGrainBatch] = useState<GrainBatchData | undefined>(grainBatchIn)
    // const [recipe, setRecipe] = useState<string | undefined>(recipeIn) // Gotten from batch serverside
    const [sizeCups, setSizeCups] = useState<number>(4) // 4 is pint!
    const [pcRun, setPcRun] = useState<PcRunData | undefined>(pcRunIn)
    const [notes, setNotes] = useState<Note[]>([])
    const [wetness, setWetness] = useState<number | undefined>(undefined) // Set on update
    const [burstGrains, setBurstGrains] = useState<number | undefined>(undefined) // Set on update

    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>()

    const cookies = useContext(CookiesContext)
    const createJar = (e: React.MouseEvent) => {
        e.preventDefault()
        if (!grainBatch) { // TODO: if recipe exists but batch does not, then create batch AND jar?
            setErr("batch must exist!")
            return
        }
        if (sizeCups < 1) {
            setErr("must select a valid jar volume")
            return
        }
        const body: any = {
            sizeCups: sizeCups,
            grainBatch: grainBatch._id,
            // optional
            wetness: wetness,
            burstGrains: burstGrains,
            pcRun: pcRun?._id, // could this be optional or required?
            notes: notes || [],
            writeTagTo: writeTagTo,
        }
        DoCreateRequest("jar", body, AssertJar, allCookies(cookies))
            .then(v=>{
                handlers.onCreate ? handlers.onCreate(new JarData(v)) : console.log("no onCreate provided")
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Create Success",
                        text: "entry created successfully",
                        isErr: false
                    }})
            })
            .catch(e=>{
                setErr("failed on createJar post: "+JSON.stringify(e))
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Create Failure",
                        text: "entry failed to create: " + JSON.stringify(e),
                        isErr: true
                    }})
            })
    }
    return <NewEntryFormWrapper entryType={"jar"} isTopLevel={handlers.isTopLevel}>
        <ErrorDisplay err={err}/>
        {grainBatchIn === undefined && <GrainBatchSelectorCloseable doSelect={setGrainBatch} allowCreation={handlers.isTopLevel} creatorInPage={handlers.isTopLevel}/>}
        <JarSizeSelector onChange={(unit: string) => {
            setSizeCups(cupsPer(unit))
        }}/>
        {pcRunIn === undefined && <PcRunSelectorCloseable doSelect={setPcRun} allowCreation={handlers.isTopLevel}
                                                          creatorInPage={handlers.isTopLevel}/>}
        <SliderOnlyIfUndefinedWithOpenButton text={"(Optional) Wetness"} defaultValue={5} onChange={setWetness}/>
        <SliderOnlyIfUndefinedWithOpenButton text={"(Optional) Burst Grains"} defaultValue={0} onChange={setBurstGrains}/>
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton"} onClick={createJar}>{"Submit new Jar"}</button>
    </NewEntryFormWrapper>
}

export function JarListPageTable({data, onClick, withLink}: ListPageItems<JarData>) {
    let cols: ListTableColumn<JarData>[] = [
        NewColumn("ID", (v) => v._id, true),
        NewColumn("Created", (v) => {
            return NumberToDateStr(v.creationDate)
        }, true),
        NewColumn("Spec", (v) => v.species || "", true),
        NewColumn("Subspec", v => v.subspecies || "", true),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: JarData) => {
            return <EntryLinkWrapper props={{entry: v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new JarData(v)}}/>
}

export function JarSelectorTable({data, onClick}: ListPageItems<JarData>) {
    return <JarListPageTable data={data} onClick={onClick}/>
}

export function JarSelector(
    {
        doSelect,
        allowCreate,
        hideDisposed
    }: {
        doSelect: (val: JarData | undefined) => void,
        allowCreate?: boolean,
        hideDisposed?:boolean
    }) {
    const table = (items: JarData[]): JSX.Element => {
        return <JarSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"jar"} entryTypes={"jars"} doSelect={doSelect} asserter={AssertJar}
                                   table={table} hideDisposed={hideDisposed}>
        {allowCreate && <NewJarForm handlers={{onCreate: doSelect, isTopLevel: false}}/>}
    </ExistingRecentSelector>
}
