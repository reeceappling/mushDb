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
    MultipartImportRequest,
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
    setFormData,
    setFormImages,
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
    AllEntries,
    OnViewCreatorQuadCol,
    SplitAllEntries
} from "@/app/components/formSubcomponents/shared";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {GenerationInput} from "@/app/components/formSubcomponents/generationInput";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {ExistingSpeciesSelector, SpeciesSubspeciesArea} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import {JarSizeSelector} from "@/app/components/formSubcomponents/utils/volumeSelector";
import {
    AclDisplay,
    IsValidAcl,
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
import WetnessSlider, {SliderOnlyIfUndefinedWithOpenButton} from "@/app/components/formSubcomponents/utils/slider";

export function AssertJar(input: any): asserts input is JarData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['recipe', 'string'],
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

export function JarImportDisplay({headerLevel}: ImportDisplayInput) {
    const [created, setCreated] = useState<number>(Date.now())
    const [recipe, setRecipe] = useState<JarRecipeData | undefined>()
    const [sizeCups, setSizeCups] = useState<number>(4)
    const [species, setSpecies] = useState<SpeciesData | undefined>()
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>()
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>()
    const [generation, setGeneration] = useState<number | undefined>()
    const [imageFile, setImageFile] = useState<File | undefined>()
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()
    const cookies = useContext(CookiesContext)
    const importEntry = () => {
        const formData = new FormData()
        if (recipe === undefined) {
            setErr("Recipe must be set!")
            return
        }
        const dataObj: any = {
            creationDate: created,
            sizeCups: sizeCups,
            recipe: recipe._id,
            // optional
            species: species?._id,
            subspecies: subspecies?._id,
            knownFruitable: knownFruitable,
            generation: generation,
            writeTagTo: writeTagTo,
        }
        if (imageFile !== undefined) {
            formData.set("img", imageFile, "img")
        }
        writeTagTo && (dataObj.writeTagTo = writeTagTo)

        MultipartImportRequest(formData, "jar", AssertJar, setErr, allCookies(cookies))
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
            <JarRecipeSelector doSelect={setRecipe} allowCreate={true}/>
        </SelectorWrapper>
        <ExistingSpeciesSelector doSelect={setSpecies}/>{/*TODO: closeable?*/}
        <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies}/>
        <KnownFruitableArea doSelect={setKnownFruitable}/>
        <GenerationInput updateParent={setGeneration}/>
        <ImageSelector updateParent={setImageFile}/>
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
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput<JarData>) {
        const [initial, setInitial] = useState(data)

        const [knownFruitable, setKnownFruitable] = useState(initial.knownFruitable)
        const [sale, setSale] = useState(initial.sale)
        const [disposed, setDisposed] = useState(initial.disposed)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [pics, setPics] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.pics))
        const [contams, setContams] = useState<SplitAllEntries<ContaminationForm, NewContaminationForm>>(InitialContamState(initial.contamination))
        const [acl, setAcl] = useState<ACL>(initial.acl)
        const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
        const [wetness, setWetness] = useState<number | undefined>(undefined)
        const [burstGrains, setBurstGrains] = useState<number | undefined>(undefined)

        // Helper states
        const [transfersOut, setTransfersOut] = useState<string[]>(initial.transfersOut || [])
        const [err, setErr] = useState<string | undefined>()
        const updateInitial = (updated: JarData) => {
            setInitial(updated)
            setKnownFruitable(updated.knownFruitable)
            setSale(updated.sale)
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
                sale: sale, // TODO: remove
                //writeTagTo: writeTagTo, // TODO: remove!
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
                setFormData(formData, dataObj)
                setFormImages(formData, "newPic", newImages)
                setFormImages(formData, "newContam", newContams)
            } catch (caught: any) {
                setErr(JSON.stringify(caught))
                return
            }

            DoUpdateMultipartRequest("jar",initial._id, formData, AssertJar, allCookies(cookies))
                .then(v=>{
                    updateInitial(new JarData(v))
                })
                .catch(e=>{
                    setErr("failed to update initial: "+JSON.stringify(e))
                })
        }
        const jarSizeArea = () => {
            return <div>
                {"Size: " + sizeFromNum(data.sizeCups)}
            </div>
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            WriteRfidOvcArea(initial._id),
        ]
        return <DisplayFormWrapper entryType={"jar"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <ID id={data._id} txt={"Grain Jar"} entryType={"jar"}/>
            <MostRecentImageDisplay data={initial.mostRecentImage} headerLevel={headerLevel}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated}
                                                disposed={disposed} setDisposedOnParent={setDisposed}
                                                readonly={readonly}/>
                    {jarSizeArea()}
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                    <JarRecipeArea headerLevel={headerLevel} recipeId={initial.recipe}/>
                    <PcRunArea binaryId={initial.pcRun} headerLevel={headerLevel}/>

                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <ParentDisplay parent={initial.parent} parentType={initial.parentType} headerLevel={headerLevel}/>
                    <InnocDisplay innoc={initial.innoc}/>
                    <KnownFruitableArea initial={knownFruitable} doSelect={setKnownFruitable} readonly={readonly}
                                        headerLevel={headerLevel}/>
                    <SaleArea sale={sale} setSale={setSale} readonly={readonly} headerLevel={headerLevel}
                              canCreateSale={true}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <GensFormDisplay gensSinceSpore={initial.genSpore} gensSinceFruitOrSpore={initial.genFruitOrSpore}
                                     headerLevel={headerLevel}/>
                </FlexedSinglesGroup>
            </FlexedArea>
            {/*TODO: validate next 2 working*/}
            {initial.wetness?<SliderOnlyIfUndefinedWithOpenButton defaultValue={5} onChange={setWetness}/> : <WetnessDisplay value={wetness} />}
            {initial.burstGrains===undefined?<SliderOnlyIfUndefinedWithOpenButton defaultValue={0} onChange={setBurstGrains}/> : <WetnessDisplay text={"Burst Grains"} value={wetness} />}

            <TransfersOutDisplay thisId={initial._id} thisEntryType={"jar"} transfersOut={transfersOut}
                                 allowNewTransferCreation={!readonly}/>
            <PicsDisplay pix={initial.pics || []} readonly={readonly} updateParent={setPics}/>
            <ContamsDisplay initial={initial.contamination || []} updateParent={setContams}
                            readonly={readonly} headerLevel={headerLevel}/>

            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay initial={acl} readonly={readonly} updateParent={setAcl}/>
            </TogglableAreaWithDepth>
            {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                e.stopPropagation();
                submit()
            }}>{"Update"}</button>}
            <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>
        </DisplayFormWrapper>
}

// NewJarForm is used from the recipe page. PcRun CAN be created from here?
// TODO: NewJarForm is used from the recipe page. PcRun CAN be created from here?
export function NewJarForm({handlers, recipeIn, pcRunIn, grainBatchIn}: {
    handlers: NewEntryInput<JarData>,
    recipeIn?: string,
    pcRunIn?: PcRunData,
    grainBatchIn?: GrainBatchData
}) {
    //const [creationDate, setCreationDate] = useState(Date.now()) // set serverside
    const [grainBatch, setGrainBatch] = useState<GrainBatchData | undefined>(grainBatchIn)
    // const [recipe, setRecipe] = useState<string | undefined>(recipeIn) // Gotten from batch serverside
    const [sizeCups, setSizeCups] = useState<number>(4) // TODO: change initial state?
    const [pcRun, setPcRun] = useState<PcRunData | undefined>(pcRunIn)
    const [notes, setNotes] = useState<Note[]>([])
    // const [wetness, setWetness] = useState(0) // Set on update
    // const [burstGrains, setBurstGrains] = useState(0) // Set on update

    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>()

    const cookies = useContext(CookiesContext)
    const createJar = (e: React.MouseEvent) => {
        e.preventDefault()
        if (!grainBatch) {
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
            // wetness: wetness,  // Added from update page
            // burstGrains: burstGrains,  // Added from update page
            pcRun: pcRun?._id, // could this be optional or required?
            notes: notes || [],
            writeTagTo: writeTagTo,
        }
        DoCreateRequest("jar", body, AssertJar, allCookies(cookies))
            .then(v=>{
                handlers.onCreate ? handlers.onCreate(v) : console.log("no onCreate provided")
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
            })
    }
    const hasGrainBatch = grainBatchIn !== undefined || recipeIn !== undefined
    return <NewEntryFormWrapper entryType={"jar"}>
        <ErrorDisplay err={err}/>
        {hasGrainBatch && <GrainBatchSelectorCloseable doSelect={setGrainBatch}
                                                               allowCreation={handlers.isTopLevel} creatorInPage={handlers.isTopLevel}/>}
        <JarSizeSelector onChange={(unit: string) => {
            setSizeCups(cupsPer(unit))
        }}/>
        {pcRunIn !== undefined && <PcRunSelectorCloseable doSelect={setPcRun} allowCreation={handlers.isTopLevel}
                                                          creatorInPage={handlers.isTopLevel}/>}
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton"} onClick={createJar}>{"Submit new Jar"}</button>
    </NewEntryFormWrapper>
}

export function JarListPageTable({data, onClick, withLink}: ListPageItems<JarData>) {
    let cols: ListTableColumn<JarData>[] = [
        NewColumn("ID", (v) => v._id),
        NewColumn("Created", (v) => {
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Spec", (v) => v.species || ""),
        NewColumn("Subspec", v => v.subspecies || ""),
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
        allowCreate
    }: {
        doSelect: (val: JarData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: JarData[]): JSX.Element => {
        return <JarSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"jar"} entryTypes={"jars"} doSelect={doSelect} asserter={AssertJar}
                                   table={table}>
        {allowCreate && <NewJarForm handlers={{onCreate: doSelect, isTopLevel: false}}/>}
    </ExistingRecentSelector>
}
