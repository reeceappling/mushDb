'use client'

import React, {JSX, useContext, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {AllEntries, OnViewCreatorQuadCol, SplitAllEntries} from "@/app/components/formSubcomponents/shared";
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
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {SaleArea} from "@/app/components/saleClient";
import {PcRunData, PcRunSelectorCloseable} from "@/app/components/pcRunServer";
import {ExistingSpeciesSelector, SpeciesSubspeciesArea} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import {SubstrateBatchArea} from "@/app/components/substrateBatchClient";
import WetnessSlider from "@/app/components/formSubcomponents/utils/slider";
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
    const ovcs: OnViewCreatorQuadCol[] = [
        OvcForNewFruit(initial._id, "bag", allCookies(cookies)), // TODO: test heavily
        WriteRfidOvcArea(initial._id),
    ]
    return (
        <DisplayFormWrapper entryType={"bag"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <ID id={data._id} txt={"Bag"} entryType={"bag"}/>
            <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>
            <MostRecentImageDisplay data={initial.mostRecentImage} headerLevel={headerLevel}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <SubstrateRecipeArea id={data.recipe} readonly={true} txt={"Substrate recipe: "}/>
                    <SubstrateBatchArea id={data.substrateBatch} txt={"Substrate batch: "} readonly={true}/>
                    <SubstrateRecipeArea id={initial.recipe} headerLevel={headerLevel} readonly={true}/>
                    {filterSizeArea(initial.filterSize)}
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <DateArea pre={"PC Date: "} readonly={true} when={initial.creationDate}/>
                    <DateArea pre={"Seal Date: "} readonly={true} when={initial.sealDate}/>
                    <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                    <DisposedDisplay disposed={disposed} setDisposedOnParent={setDisposed} readonly={readonly}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <KnownFruitableArea initial={initial.knownFruitable} doSelect={setKnownFruitable}
                                        readonly={readonly}
                                        headerLevel={headerLevel}/>
                    <SaleArea sale={initial.sale} setSale={setSale} readonly={readonly}
                              headerLevel={headerLevel} canCreateSale={true}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <GensFormDisplay gensSinceSpore={initial.genSpore}
                                     gensSinceFruitOrSpore={initial.genFruitOrSpore}/>
                    <WetnessDisplay value={initial.wetness}/>{/* TODO: FIX */}
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                    {/*TODO: THIS!<SpeciesSubspeciesFormArea species={initial.species} subspecies={initial.subspecies}/>*/}
                    <ParentDisplay parent={initial.parent} parentType={initial.parentType}
                                   headerLevel={headerLevel}/>
                    <InnocDisplay innoc={initial.innoc}/>
                </FlexedSinglesGroup>
            </FlexedArea>

            <TransfersOutDisplay validTypesTo={["plate"]} thisId={initial._id} thisEntryType={"bag"}
                                 transfersOut={initial.transfersOut}
                                 allowNewTransferCreation={true}/>
            <PicsDisplay pix={initial.pics || []} readonly={readonly} updateParent={setPics}/>{/* Pics */}
            {/* Flushes */}
            <PicsDisplay pix={initial.flushes || []} readonly={readonly}
                         updateParent={setFlushes} addButtonText={"Create New Flush"}
                         sectionHeader={"Flushes: "}/>

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
    return <TestAndValidate todos={["fix"]}>
        <div>{(text||"Wetness")+": " + (value ? value + "/10" : "unknown")}</div>
    </TestAndValidate>
}

const filterSizeSelector = (setFilterSize: (f?: string) => void, filterSize?: string) => {
    return <div className={"centerH medGapTop"}>
        {"Filter size: "}<FilterSizeSelector onSelect={setFilterSize} current={filterSize}/>{/* TODO: ensure working!*/}
    </div>
}

export function NewBagForm({handlers, substrateBatchIn, pcRunIn}: {
    handlers: NewEntryInput<BagData>,
    pcRunIn?: PcRunData,
    substrateBatchIn?: SubstrateBatchData,
}) {
    // Required defined
    const [substrateBatch, setSubstrateBatch] = useState(substrateBatchIn)
    const [wetness, setWetness] = useState<number | undefined>(undefined)
    const [pcRun, setPcRun] = useState<PcRunData | undefined>(pcRunIn)
    const [filterSize, setFilterSize] = useState<string | undefined>()
    // Optional
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
                handlers.onCreate ? handlers.onCreate(v) : console.log("no onCreate provided")
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
                <SubstrateBatchSelectorCloseable txt={"Substrate Batch (FIXME)"} doSelect={setSubstrateBatch}
                                                 allowCreation={handlers.isTopLevel} creatorInPage={false}/>}
            <WetnessSlider defaultValue={5} onChange={(event: Event, value: number, activeThumb: number) => {
                setWetness(value)
            }}/>
            {pcRunIn === undefined && // TODO: Show pc run if already exists?
                <PcRunSelectorCloseable doSelect={setPcRun} creatorInPage={true} allowCreation={true}/>}
            <div className={"centerH medGapTop"}>
                {"Filter size: "}<FilterSizeSelector onSelect={setFilterSize}
                                                     current={filterSize}/>{/* TODO: ensure working!*/}
            </div>
            <NewEntryNotes setNotes={setNotes}/>
            {/* Write tag area */}
            <ReaderWriterSelector txt={"Write to: "} onSelect={setWriteTagTo}/>
            {/* SUBMIT AREA */}
            <input type="submit" value="Submit" className={"bottomButton"} onClick={newBagSubmit} onSubmit={(e) => {
                e.preventDefault();
            }}/>
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
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>(undefined)
    const [generation, setGeneration] = useState<number | undefined>(undefined)
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
            subspecies: subspecies?._id,
            generation: generation,
            knownFruitable: knownFruitable,
            writeTagTo: writeTagTo,
        }
        setFormData(formData, dataObj)
        imageFile && formData.set("img", imageFile, "img")
        MultipartImportRequest(formData, "bag", AssertBag, setErr, allCookies(cookies))
    }
    return <ImportEntryFormWrapper entryType={"bag"}>
        {/* Required Fields */}
        <ErrorDisplay err={err} headerLevel={headerLevel}/>
        <DateArea pre={"Seal Date: "} when={Date.now()} updateParent={setSealDate}/>
        <SelectorWrapper current={recipe} title={"Recipe"} nameFunc={(v: SubstrateRecipeData) => v._id}>
            <SubstrateRecipeSelector doSelect={setRecipe} allowCreate={false} creatorInPage={false}/>
        </SelectorWrapper>
        <div className={"centerH medGapTop"}>
            {"Filter size: "}<FilterSizeSelector onSelect={setFilterSize}
                                                 current={filterSize}/>{/* TODO: ensure working!*/}
        </div>
        <ExistingSpeciesSelector doSelect={setSpecies}/>

        {/* Optional fields*/}
        <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies}/>
        <GenerationInput updateParent={setGeneration}/>
        <KnownFruitableArea doSelect={setKnownFruitable}/>

        <TopLevelImageSelector updateParent={setImageFile} buttonText={"Upload image"}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button onClick={submitImportBag} className={"greenButton bottomButton"}>{"Submit"}</button>
    </ImportEntryFormWrapper>
}

export function BagListPageTable({data, onClick, withLink}: ListPageItems<BagData>) {
    let cols: ListTableColumn<BagData>[] = [
        NewColumn("ID", (v) => v._id),
        NewColumn("Created", (v) => {
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
        NewColumn("Species", (v) => v.species || ""),
        NewColumn("Subspec.", (v) => v.subspecies || ""),
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
