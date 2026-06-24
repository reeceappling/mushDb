'use client'

import React, {JSX, useContext, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {
    AddCreatedTriColFunction,
    AllEntries,
    OnViewCreatorQuadCol,
    SplitAllEntries
} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {BagData} from "@/app/components/bagServer";
import {
    InitialPicsEntries,
    IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
    PicWithNotesForm,
} from "@/app/components/formSubcomponents/picWithNotes";
import {InnocDisplay, TransfersOutDisplay} from "@/app/components/transferClient";
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
    setFormData,
    setFormImages,
} from "@/app/components/common";
import {
    DisposedDisplay,
    ErrorDisplay,
    GensFormDisplay,
    MostRecentImageDisplay,
    ParentDisplay,
    PicsDisplay
} from "@/app/components/formSubcomponents/commonClient";
import {KnownFruitableArea} from "@/app/components/formSubcomponents/knownFruitableArea";
import {SubstrateRecipeArea, SubstrateRecipeSelector} from "@/app/components/substrateRecipeClient";
import {
    ContaminationForm,
    ContamsDisplay,
    InitialContamState,
    IsValidContamination,
    NewContaminationForm
} from "@/app/components/formSubcomponents/contaminations";
import {GenerationInput} from "@/app/components/formSubcomponents/generationInput";
import {TopLevelImageSelector} from "@/app/components/formSubcomponents/imageSelector";
import ReaderWriterSelector, {
    WriteRfidOvcArea
} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {SubstrateRecipeData} from "@/app/components/substrateRecipeServer";
import {SpeciesData} from "@/app/components/speciesServer";
import {SaleArea} from "@/app/components/saleClient";
import {PcRunData, PcRunSelectorCloseable} from "@/app/components/pcRunServer";
import {
    ExistingSpeciesSubspeciesSelector,
    SpeciesSubspeciesArea
} from "@/app/components/speciesClient";
import {SubstrateBatchArea} from "@/app/components/substrateBatchClient";
import {SliderOnlyIfUndefinedWithOpenButton} from "@/app/components/formSubcomponents/utils/slider";
import {SubstrateBatchData, SubstrateBatchSelectorCloseable} from "@/app/components/substrateBatchServer";
import TestAndValidate from "@/app/components/testing/untested";
import {AclDisplay, MarshalAcl, TogglableAreaWithDepth, UnmarshalAcl} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {OnViewCreatorsQuadColArea, OvcForNewFruit} from "@/app/components/formSubcomponents/ovc";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {useQuery} from "@tanstack/react-query";
import {GetFilterSizes} from "@/app/components/formSubcomponents/server";
import {SelectorFor} from "@/app/components/selector";
import {PcRunArea} from "@/app/components/pcRunClient";

export function AssertBag(input: any): asserts input is BagData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['recipe', 'string'],
        ['filterSize', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Bag assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // optional simple keys
    const optionalSimpleKeys = new Map<string, string>([
        ['pcRun', 'string'],
        ['genSpore', 'number'],
        ['genFruitOrSpore', 'number'],
        ['sealDate', 'number'],
        ['knownFruitable', 'boolean'],
        ['species', 'string'],
        ['subspecies', 'string'],
        ['innoc', 'string'],
        ['parentType', 'string'],
        ['parent', 'string'],
        ['sale', 'string'],
        ['substrateBatch', 'string'],
        ['disposed', 'number'],
        ['wetness', 'number'],
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
            throw new Error('Bag assertion failure: required key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    const complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
    ])
    for (const [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Bag assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', (item) => {
            return typeof item === 'string'
        }],
        ['pics', IsValidPicWithNotesIncoming],
        ['contamination', IsValidContamination],
        ['flushes', IsValidPicWithNotesIncoming],
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

export default function BagDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput<BagData>) {
    const [initial, setInitial] = useState(data)

    const [knownFruitable, setKnownFruitable] = useState(initial.knownFruitable)
    const [sale, setSale] = useState(initial.sale)
    const [transfersOut, setTransfersOut] = useState(initial.transfersOut || [])
    const [disposed, setDisposed] = useState(initial.disposed)
    const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(data.notes))
    const [wetness, setWetness] = useState(initial.wetness) // TODO: new! handle on go side!
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    // ItemsWithPics
    const [pics, setPics] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.pics))
    const [contams, setContams] = useState<SplitAllEntries<ContaminationForm, NewContaminationForm>>(InitialContamState(initial.contamination))
    const [flushes, setFlushes] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.flushes))
    const [err, setErr] = useState<string | undefined>()
    const [acl, setAcl] = useState<ACL>(initial.acl)
    //const [newFruits, setNewFruits] = useState<FruitData[]>([]) // TODO: get rid of???
    const filterSizeArea = (filterSize: string, headerLevel?: number) => {
        return <div>
            <div>{"Filter Size: " + filterSize}</div>
        </div>
    }
    const updateInitial = (updated: BagData) => {
        setInitial(updated)
        setWetness(updated.wetness)
        setKnownFruitable(updated.knownFruitable)
        setSale(updated.sale)
        setTransfersOut(updated.transfersOut || [])
        setDisposed(updated.disposed)
        setNotes(InitialNotesState(updated.notes))
        setPics(InitialPicsEntries(updated.pics))
        setContams(InitialContamState(updated.contamination))
        setFlushes(InitialPicsEntries(updated.flushes))
        setAcl(updated.acl)
        setErr(undefined)
    }
    const cookies = useContext(CookiesContext)
    const bagSubmit = () => {
        const formData = new FormData()
        const dataObj: any = {
            wetness: wetness, // TODO: ok?
            knownFruitable: knownFruitable,
            sale: sale, // TODO: how/when should sales be made?
            disposed: disposed,
            notes: notes,
            writeTagTo: writeTagTo,
            acl: MarshalAcl(acl),
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
            // Flushes
            const flushesInfo = resolvePicsFormData(flushes)
            const newFlushes = flushesInfo.images
            dataObj.flushes = flushesInfo.obj
            // Set data on form
            setFormData(formData, dataObj)
            //formData.set("data", JSON.stringify(dataObj))
            setFormImages(formData, "newPic", newImages)
            setFormImages(formData, "newContam", newContams)
            setFormImages(formData, "newFlush", newFlushes)
        } catch (caught: any) {
            setErr(JSON.stringify(caught))
            return
        }
        DoUpdateMultipartRequest("bag", initial._id, formData, AssertBag, allCookies(cookies))
            .then(v => {
                updateInitial(new BagData(v))
            })
            .catch(e => {
                setErr("failed to update initial: " + JSON.stringify(e))
            })
    }
    const isInnoculated = ()=>{
        return initial.species !== undefined
    }
    const ovcs:()=>OnViewCreatorQuadCol[] = ()=>{
        const innoculated = isInnoculated()
        const disp = initial.disposed !== undefined
        return (!disp)?[
            ...((innoculated)?[OvcForNewFruit(initial._id, "bag", allCookies(cookies))]:[]),
            WriteRfidOvcArea(initial._id), // TODO: TEST!
            ...((innoculated)?[
                {
                    txt: "Create Spore Print (+fruit)",
                    newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                        return <TestAndValidate todos={["not implemented yet, should also create fruit!", "Do MUCH later. Shortcut"]}>
                            <div>{"Not yet implemented!"}</div>
                        </TestAndValidate>
                    },
                    needsTesting: true,
                },
                {
                    txt: "Create Spore Swab (+fruit)",
                    newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                        return <TestAndValidate todos={["not implemented yet, should also create fruit!", "Do MUCH later. Shortcut"]}>
                            <div>{"Not yet implemented!"}</div>
                        </TestAndValidate>
                    },
                    needsTesting: true,
                }
            ]:[])
        ]:[]
    }
    return (
        <DisplayFormWrapper entryType={"bag"}>
            <ErrorDisplay err={err}/>
            <ID props={{id:data._id, txt:"Bag", entryType:"bag"}}/>
            <OnViewCreatorsQuadColArea OnViewCreators={ovcs()} readonly={readonly}/>
            <MostRecentImageDisplay data={initial.mostRecentImage}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <PcRunArea binaryId={initial.pcRun}/>{/* TODO: ENSURE OK!*/}
                    <SubstrateRecipeArea id={data.recipe} readonly={true} txt={"Substrate recipe: "}/>{/* TODO: hover styling?*/}
                    <SubstrateBatchArea id={data.substrateBatch} txt={"Substrate batch: "}/>{/* TODO: hover styling?*/}
                    {filterSizeArea(initial.filterSize)}
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <DateArea pre={"PC Date: "} readonly={true} when={initial.creationDate}/>
                    <DateArea pre={"Seal Date: "} readonly={true} when={initial.sealDate}/>
                    <DateArea pre={"Last Updated: "} readonly={true} when={initial.lastUpdated}/>
                    <DisposedDisplay initial={initial.disposed} setDisposedOnParent={setDisposed} readonly={readonly}/>
                </FlexedSinglesGroup>
                {isInnoculated()&&<FlexedSinglesGroup>
                    <KnownFruitableArea initial={initial.knownFruitable} doSelect={setKnownFruitable}
                                        readonly={readonly}
                                        headerLevel={headerLevel}/>
                    <SaleArea sale={initial.sale} setSale={setSale} readonly={readonly} canCreateSale={true}/>
                </FlexedSinglesGroup>}
                <FlexedSinglesGroup>
                    {isInnoculated()&&<GensFormDisplay gensSinceSpore={initial.genSpore} gensSinceFruitOrSpore={initial.genFruitOrSpore}/>}
                    {initial.wetness ? <WetnessDisplay value={initial.wetness}/> : <SliderOnlyIfUndefinedWithOpenButton text={"Wetness"} defaultValue={5} onChange={setWetness}/>}
                </FlexedSinglesGroup>
                {isInnoculated()&&<FlexedSinglesGroup>
                    <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                    <ParentDisplay parent={initial.parent} parentType={initial.parentType}/>
                    <InnocDisplay innoc={initial.innoc}/>
                </FlexedSinglesGroup>}
            </FlexedArea>

            {isInnoculated()&&<TransfersOutDisplay thisId={initial._id} thisEntryType={"bag"} /*validTypesTo={["plate"]} TODO: on go side*/
                                 transfersOut={initial.transfersOut} requireConfirmation={true/* TODO: ok?*/}
                                 allowNewTransferCreation={true}/>}
            <PicsDisplay pix={initial.pics || []} readonly={readonly} updateParent={setPics}/>{/* Pics */}
            {/* Flushes */}
            {isInnoculated()&&<PicsDisplay pix={initial.flushes || []} readonly={readonly}
                         updateParent={setFlushes} addButtonText={"Create New Flush"}
                         sectionHeader={"Flushes: "}/>}

            <ContamsDisplay initial={initial.contamination || []} updateParent={setContams}
                            readonly={readonly} headerLevel={headerLevel}/>
            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay initial={acl} readonly={readonly} updateParent={setAcl}/>
            </TogglableAreaWithDepth>
            {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                e.stopPropagation();
                bagSubmit()
            }}>{"Update"}</button>}
        </DisplayFormWrapper>
    )
}

export function WetnessDisplay({value,text}: { value?: number, text?: string}) {
    return <div>{(text||"Wetness")+": " + (value ? value + "/10" : "unknown")}</div>
}

// const filterSizeSelector = (setFilterSize: (f?: string) => void, filterSize?: string) => {
//     return <div className={"centerH medGapTop"}>
//         {"Filter size: "}<FilterSizeSelector onSelect={setFilterSize} current={filterSize}/>{/* TODO: ensure working!*/}
//     </div>
// }

export function NewBagForm({handlers, substrateBatchIn, pcRunIn}: {
    handlers: NewEntryInput<BagData>,
    pcRunIn?: PcRunData,
    substrateBatchIn?: SubstrateBatchData,
}) {
    // Required defined
    const [substrateBatch, setSubstrateBatch] = useState(substrateBatchIn)
    const [pcRun, setPcRun] = useState<PcRunData | undefined>(pcRunIn)
    const [filterSize, setFilterSize] = useState<string | undefined>()
    // Optional
    const [wetness, setWetness] = useState<number | undefined>(undefined) // TODO: allow to be undefined on go side
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()
    const cookies = useContext(CookiesContext)
    const newBagSubmit = () => {
        if (pcRun === undefined) {
            setErr("PC Run cannot be undefined!");
            return
        }
        if (filterSize === undefined) {
            setErr("Filter Size cannot be undefined!");
            return
        }
        if (substrateBatch === undefined) {
            setErr("Substrate batch cannot be undefined!");
            return
        }
        const body: any = {
            substrateBatch: substrateBatch._id,
            wetness: wetness,
            pcRun: pcRun._id,
            filterSize: filterSize,
            notes: notes,
            writeTagTo: writeTagTo,
        }
        DoCreateRequest("bag", body, AssertBag, allCookies(cookies))
            .then(v => {
                handlers.onCreate ? handlers.onCreate(new BagData(v)) : console.log("no onCreate provided")
            })
            .catch(e => {
                setErr(JSON.stringify(e))
            })
    }
    return (
        <NewEntryFormWrapper entryType={"bag"}>
            <ErrorDisplay err={err}/>
            <div>{"Creating Bag: "}</div>
            {substrateBatchIn !== undefined &&
                <SubstrateBatchSelectorCloseable txt={"Substrate Batch"} doSelect={setSubstrateBatch}
                                                 allowCreation={handlers.isTopLevel} creatorInPage={false}/>}
            <SliderOnlyIfUndefinedWithOpenButton defaultValue={5} onChange={setWetness}/>
            {/*<WetnessSlider defaultValue={5} onChange={(event: Event, value: number, activeThumb: number) => {*/}
            {/*    setWetness(value)*/}
            {/*}}/>*/}
            {pcRunIn === undefined && // TODO: Show pc run if already exists?
                <PcRunSelectorCloseable doSelect={setPcRun} creatorInPage={true} allowCreation={true}/>}
            <div className={"centerH medGapTop"}>
                {"Filter size: "}<FilterSizeSelector onSelect={setFilterSize}
                                                     current={filterSize}/>
            </div>
            <NewEntryNotes setNotes={setNotes}/>
            {/* Write tag area */}
            <ReaderWriterSelector txt={"Write to: "} onSelect={setWriteTagTo}/>
            {/* SUBMIT AREA */}
            <button className={"greenButton buttonFullWidth"} onClick={e=>{
                e.stopPropagation()
                newBagSubmit()
            }}>{"Create New Bag"}</button>
        </NewEntryFormWrapper>
    )
}

export function BagImportDisplay({headerLevel}: ImportDisplayInput) {
    // Required
    const [sealDate, setSealDate] = useState(Date.now())
    const [species, setSpecies] = useState<SpeciesData | undefined>(undefined)
    const [recipe, setRecipe] = useState<SubstrateRecipeData | undefined>(undefined)
    const [filterSize, setFilterSize] = useState<string | undefined>(undefined)
    // Optional
    const [subspecies, setSubspecies] = useState<string | undefined>(undefined)
    const [generation, setGeneration] = useState<number | undefined>(1)
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>(undefined)
    const [imageFile, setImageFile] = useState<File | undefined>(undefined)
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const cookies = useContext(CookiesContext)
    const submitImportBag = () => {
        const reqd = new Map<string, any>([
            ['recipe', recipe],
            ['filterSize', filterSize],
        ])
        for (const [key, val] of reqd) {
            if (!val) {
                setErr(key + " must be defined!");
                return
            }
        }
        const formData = new FormData()
        const dataObj: any = {
            creationDate: sealDate,
            recipe: recipe?._id, // MUST EXIST
            filterSize: filterSize,
            // optional
            species: species?._id,
            subspecies: subspecies,
            generation: generation,
            knownFruitable: knownFruitable,
            writeTagTo: writeTagTo,
        }
        setFormData(formData, dataObj)
        imageFile && formData.set("img", imageFile, "img")
        DoMultipartImportRequest(formData, "bag", AssertBag, setErr, allCookies(cookies))
        // TODO: redirect not working.

    }
    return <ImportEntryFormWrapper entryType={"bag"}>
        {/* Required Fields */}
        <ErrorDisplay err={err}/>
        <DateArea pre={"Seal Date: "} when={Date.now()} updateParent={setSealDate}/>
        <SelectorWrapper current={recipe} title={"Recipe"} nameFunc={(v: SubstrateRecipeData) => v._id}>
            <SubstrateRecipeSelector doSelect={setRecipe} allowCreate={false} creatorInPage={false}/>
        </SelectorWrapper>
        <div className={"centerH medGapTop"}>
            {"Filter size: "}<FilterSizeSelector onSelect={setFilterSize}
                                                 current={filterSize}/>{/* TODO: ensure working!*/}
        </div>
        {/* TODO: WETNESS*/}
        {/* TODO: NOTES*/}
        {/* Species required, subspecies optional*/}
        <ExistingSpeciesSubspeciesSelector doSelectSpecies={setSpecies} doSelectSubspecies={setSubspecies}/>
        {/*<ExistingSpeciesSelector doSelect={setSpecies}/>*/}
        {/*<TestAndValidate todos={["what to do when a species has no subspecies?"]}>*/}
        {/*    <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies}/>*/}
        {/*</TestAndValidate>*/}
        {/* Other Optional fields*/}
        {species && <><GenerationInput updateParent={setGeneration}/>
            <TestAndValidate todos={["default to unknown?"]}>
                <KnownFruitableArea doSelect={setKnownFruitable}/>
            </TestAndValidate></>}

        <TopLevelImageSelector updateParent={setImageFile} buttonText={"Upload image"}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button onClick={submitImportBag} className={"greenButton bottomButton"}>{"Submit"}</button>
    </ImportEntryFormWrapper>
}

export function BagListPageTable({data, onClick, withLink}: ListPageItems<BagData>) {
    let cols: ListTableColumn<BagData>[] = [
        NewColumn("ID", (v) => v._id, true),
        NewColumn("Created", (v) => {
            return NumberToDateStr(v.creationDate)
        }, true),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }, true),
        NewColumn("Species", (v) => v.species || "", true),
        NewColumn("Subspec.", (v) => v.subspecies || ""), // TODO: fit?
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: BagData) => {
            return <EntryLinkWrapper props={{entry: v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v => {
        return new BagData(v)
    }}/>
}

export function BagSelectorTable({data, onClick}: ListPageItems<BagData>) {
    return <BagListPageTable data={data} onClick={onClick} withLink={true}/>
}

export function BagSelector(
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: BagData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: BagData[]): JSX.Element => {
        return <BagSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"bag"} entryTypes={"bags"} doSelect={doSelect} asserter={AssertBag}
                                   table={table}>
        {allowCreate && <NewBagForm handlers={{onCreate: doSelect, isTopLevel: false}}/>}
    </ExistingRecentSelector>
}

export function FilterSizeSelector(
    {current, onSelect}: {
        current?: string,
        onSelect?: (ab?: string) => void
    }) {
    const {isPending, error, data} = useQuery({
        queryKey: ['filterSizes'],
        queryFn: GetFilterSizes,
    })
    if (isPending || error !== null) {
        return <div>{isPending ? "FILTER SIZE SELECTOR LOADING" : "FILTER SIZE SELECTOR ERROR: " + error.message}</div>
    }
    return <SelectorFor disabled={onSelect === undefined} options={["", ...data.keys()]} initial={current || ""}
                        updateParent={(s) => {
                            if (s === "") {
                                onSelect && onSelect(undefined)
                            }
                            onSelect && onSelect(s as string)
                        }
                        }/>
}
